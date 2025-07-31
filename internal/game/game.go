package game

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"quards/internal/database"
	"quards/internal/deck"
	"quards/internal/lens"
	"quards/internal/parser"
)

// writeEventLog creates a properly formatted event-sourcing log entry
func writeEventLog(event string, params map[string]interface{}) string {
	line := event
	for key, value := range params {
		// Quote string values to handle spaces properly
		if str, ok := value.(string); ok {
			line += fmt.Sprintf(" %s=%q", key, str)
		} else {
			line += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	return line
}

// mapActionToEventName maps old action names to new event names
func mapActionToEventName(action string) string {
	switch action {
	case "draw_card":
		return "CardDrawn"
	case "ink_card":
		return "CardInked"
	case "play_card":
		return "CardPlayed"
	case "quest":
		return "QuestAttempted"
	case "pass":
		return "TurnPassed"
	case "turn_start":
		return "TurnStarted"
	default:
		return action // Fallback
	}
}

// Game represents a game instance
type Game struct {
	ID          int       `json:"id"`
	Player1Deck string    `json:"player1Deck"`
	Player2Deck string    `json:"player2Deck"`
	Seed        *int      `json:"seed"`
	LogContent  string    `json:"logContent"`
	Status      string    `json:"status"`
	Winner      *int      `json:"winner"`
	Turns       int       `json:"turns"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// GameList represents a simplified game list for API responses
type GameList struct {
	ID          int       `json:"id"`
	Player1Deck string    `json:"player1Deck"`
	Player2Deck string    `json:"player2Deck"`
	Seed        *int      `json:"seed"`
	Status      string    `json:"status"`
	Winner      *int      `json:"winner"`
	Turns       int       `json:"turns"`
	Created     time.Time `json:"created"`
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	Player1Deck string `json:"player1Deck"` // Can be deck name or ID as string
	Player2Deck string `json:"player2Deck"` // Can be deck name or ID as string  
	Seed        *int   `json:"seed"`
	LogContent  string `json:"logContent,omitempty"` // For uploaded games
}

// resolveDeckName resolves a deck identifier (name or ID as string) to a deck name
func resolveDeckName(deckIdentifier string) (string, error) {
	// Try to parse as integer ID first
	if id, err := strconv.Atoi(deckIdentifier); err == nil {
		// It's an ID, load deck by ID and return name
		deckData, err := deck.LoadDeckByID(id)
		if err != nil {
			return "", fmt.Errorf("failed to load deck by ID %d: %w", id, err)
		}
		return deckData.Name, nil
	}
	
	// It's a name, validate it exists
	_, err := deck.LoadDeck(deckIdentifier)
	if err != nil {
		return "", fmt.Errorf("failed to load deck by name %s: %w", deckIdentifier, err)
	}
	
	return deckIdentifier, nil
}

// CreateGame creates a new game and generates initial log
func CreateGame(req *CreateGameRequest) (*Game, error) {
	db := database.GetDB()

	// Resolve deck identifiers to names
	player1DeckName, err := resolveDeckName(req.Player1Deck)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve player 1 deck: %w", err)
	}
	
	player2DeckName, err := resolveDeckName(req.Player2Deck)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve player 2 deck: %w", err)
	}

	// Generate seed if not provided
	seed := req.Seed
	if seed == nil {
		generatedSeed := rand.Intn(1000000)
		seed = &generatedSeed
	}

	// Use provided log content or generate initial log
	var logContent string
	if req.LogContent != "" {
		logContent = req.LogContent
	} else {
		logContent = generateInitialLog(player1DeckName, player2DeckName, *seed)
	}

	// Insert game into database
	var gameID int
	err = db.QueryRow(`
		INSERT INTO games (player1_deck, player2_deck, seed, log_content, status)
		VALUES ($1, $2, $3, $4, 'created')
		RETURNING id`,
		player1DeckName, player2DeckName, seed, logContent).Scan(&gameID)

	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Load and return the created game
	return LoadGame(gameID)
}

// LoadGame loads a game by ID
func LoadGame(id int) (*Game, error) {
	db := database.GetDB()

	var game Game
	err := db.QueryRow(`
		SELECT id, player1_deck, player2_deck, seed, log_content, 
		       status, winner, turns, created_at, modified_at
		FROM games WHERE id = $1`, id).Scan(
		&game.ID, &game.Player1Deck, &game.Player2Deck,
		&game.Seed, &game.LogContent, &game.Status, &game.Winner,
		&game.Turns, &game.Created, &game.Modified)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game not found: %d", id)
		}
		return nil, fmt.Errorf("failed to load game: %w", err)
	}

	return &game, nil
}

// LoadGameByName loads a game by name
func LoadGameByName(name string) (*Game, error) {
	db := database.GetDB()

	var game Game
	err := db.QueryRow(`
		SELECT id, player1_deck, player2_deck, seed, log_content,
		       status, winner, turns, created_at, modified_at
		FROM games WHERE name = $1`, name).Scan(
		&game.ID, &game.Player1Deck, &game.Player2Deck,
		&game.Seed, &game.LogContent, &game.Status, &game.Winner,
		&game.Turns, &game.Created, &game.Modified)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game not found: %s", name)
		}
		return nil, fmt.Errorf("failed to load game: %w", err)
	}

	return &game, nil
}

// ListGames returns a list of all games, optionally filtered by deck
func ListGames(deckFilter string) ([]GameList, error) {
	db := database.GetDB()

	query := `
		SELECT id, player1_deck, player2_deck, seed, status, 
		       winner, turns, created_at
		FROM games`
	args := []interface{}{}

	if deckFilter != "" {
		query += " WHERE player1_deck = $1 OR player2_deck = $1"
		args = append(args, deckFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []GameList
	for rows.Next() {
		var game GameList
		err := rows.Scan(&game.ID, &game.Player1Deck, &game.Player2Deck,
			&game.Seed, &game.Status, &game.Winner, &game.Turns, &game.Created)
		if err != nil {
			continue // Skip invalid rows
		}
		games = append(games, game)
	}

	return games, nil
}

// DeleteGame removes a game from the database
func DeleteGame(id int) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM games WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("game not found: %d", id)
	}

	return nil
}

// generateInitialLog creates a basic game log with setup actions in JSON Lines format
func generateInitialLog(player1DeckName, player2DeckName string, seed int) string {
	var entries []string

	// Load deck contents
	player1DeckData, err := deck.LoadDeck(player1DeckName)
	if err != nil {
		return "Failed to load player one's deck: " + err.Error()
	}

	player2DeckData, err := deck.LoadDeck(player2DeckName)
	if err != nil {
		return "Failed to load player two's deck: " + err.Error()
	}

	// Create card lists from deck data
	player1Cards := expandDeckToCards(player1DeckData.Cards)
	player2Cards := expandDeckToCards(player2DeckData.Cards)

	// Shuffle using the provided seed
	rng := rand.New(rand.NewSource(int64(seed)))
	rng.Shuffle(len(player1Cards), func(i, j int) {
		player1Cards[i], player1Cards[j] = player1Cards[j], player1Cards[i]
	})
	rng.Shuffle(len(player2Cards), func(i, j int) {
		player2Cards[i], player2Cards[j] = player2Cards[j], player2Cards[i]
	})

	// Game start entry
	entries = append(entries, writeEventLog("GameStarted", map[string]interface{}{
		"p1_deck": player1DeckName,
		"p2_deck": player2DeckName,
		"seed":    seed,
	}))

	// Shuffle decks
	entries = append(entries, writeEventLog("DecksShuffled", map[string]interface{}{
		"seed": seed,
	}))

	// Technically this is derived data. We can compute the state of the deck, and we
	// should know that the start of the game includes drawing the first 7 cards.
	// Explore dropping this from the logs if it's problematic.
	// Draw opening hands - single action with all cards for both players
	p1CardsStr := strings.Join(player1Cards[:7], ",")
	p2CardsStr := strings.Join(player2Cards[:7], ",")
	entries = append(entries, writeEventLog("OpeningHandsDrawn", map[string]interface{}{
		"p1": fmt.Sprintf("\"%s\"", p1CardsStr),
		"p2": fmt.Sprintf("\"%s\"", p2CardsStr),
	}))

	// Player 1 starts (turn 1)
	entries = append(entries, writeEventLog("TurnStarted", map[string]interface{}{
		"player": 1,
		"turn":   1,
	}))

	return strings.Join(entries, "\n") + "\n"
}
func AppendActionToGame(gameName, actionType string, parameters map[string]interface{}) error {
	// Load the game
	gameData, err := LoadGameByName(gameName)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	// Parse current log to get the current state
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		return fmt.Errorf("failed to parse game log: %w", err)
	}

	// Use lens to get current game state instead of manual parsing
	processor := lens.New()
	gameStateData, err := processor.Lens("gameState", entries)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}
	
	gameState := gameStateData.(map[string]interface{})
	currentPlayer := gameState["currentPlayer"].(int)
	currentTurn := gameState["currentTurn"].(int)

	// Create the new log entry in event-sourcing format
	eventName := mapActionToEventName(actionType)
	// Add player parameter if not already present
	if currentPlayer > 0 {
		parameters["player"] = currentPlayer
	}
	newLogLine := writeEventLog(eventName, parameters)

	// Start building the updated log content
	updatedLogContent := gameData.LogContent + newLogLine + "\n"

	// If this is a pass action, we need to add automatic follow-up actions
	if actionType == "pass" {
		// Get the next player (who will start their turn)
		nextPlayer := currentPlayer
		if currentPlayer == 1 {
			nextPlayer = 2
		} else {
			nextPlayer = 1
		}

		// Determine the next turn number
		nextTurn := currentTurn + 1

		// Player 1 doesn't draw on turn 1 (their very first turn), but draws on all other turns
		// Player 2 always draws at the start of their turns
		shouldDraw := !(nextTurn == 1 && nextPlayer == 1)

		// Add draw_card action if the player should draw
		if shouldDraw {
			// Get next card from deck for the player
			nextCard, err := getNextCardFromDeck(entries, gameData, nextPlayer)
			if err == nil && nextCard != "" {
				drawLogLine := writeEventLog("CardDrawn", map[string]interface{}{
					"card_id": nextCard,
					"player":  nextPlayer,
				})
				updatedLogContent += drawLogLine + "\n"
			}
		}

		// Add turn_start action for the next player
		turnStartLogLine := writeEventLog("TurnStarted", map[string]interface{}{
			"player": nextPlayer,
			"turn":   nextTurn,
		})
		updatedLogContent += turnStartLogLine + "\n"
	}

	// Update the game in the database
	db := database.GetDB()
	// TODO: Consider using a different storage mechanism for this (appending instead of replacing))
	_, err = db.Exec("UPDATE games SET log_content = $1, modified_at = NOW() WHERE name = $2",
		updatedLogContent, gameName)
	if err != nil {
		return fmt.Errorf("failed to update game log: %w", err)
	}

	return nil
}

// getNextCardFromDeck determines what card a player should draw next from their deck
func getNextCardFromDeck(entries []parser.LogEntry, gameData *Game, player int) (string, error) {
	// First, find the game_start entry to get deck names and seed
	var player1DeckName, player2DeckName string
	var seed int

	for _, entry := range entries {
		if string(entry.Event) == "GameStarted" && entry.Parameters != nil {
			if deck1, ok := entry.Parameters["p1_deck"]; ok {
				player1DeckName = deck1
				fmt.Printf("DEBUG DECK: p1_deck='%s'\n", player1DeckName)
			}
			if deck2, ok := entry.Parameters["p2_deck"]; ok {
				player2DeckName = deck2
				fmt.Printf("DEBUG DECK: p2_deck='%s'\n", player2DeckName)
			}
			seed = entry.GetInt("seed")
		}
	}

	if player1DeckName == "" || player2DeckName == "" {
		return "", fmt.Errorf("could not find deck names in game log")
	}

	// Load the appropriate deck
	var deckName string
	if player == 1 {
		deckName = player1DeckName
	} else {
		deckName = player2DeckName
	}

	deckData, err := deck.LoadDeck(deckName)
	if err != nil {
		return "", fmt.Errorf("failed to load deck %s: %w", deckName, err)
	}

	// Expand deck to card list and shuffle with the same seed
	cards := expandDeckToCards(deckData.Cards)
	rng := rand.New(rand.NewSource(int64(seed)))
	rng.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	// Track cards already drawn/used by this player
	var cardsUsed []string

	// Process all entries to track card movements
	for _, entry := range entries {
		if entry.GetPlayer() == player {
			switch string(entry.Event) {
			case "CardDrawn", "CardInked", "CardPlayed":
				if entry.Parameters != nil {
					cardID := entry.GetCard("card_id")
					if cardID != "" {
						cardsUsed = append(cardsUsed, cardID)
					}
				}
			}
		} else if string(entry.Event) == "OpeningHandsDrawn" && entry.Parameters != nil {
			// Handle opening hands
			var playerCards []interface{}
			if player == 1 {
				cardStrings := entry.GetStringSlice("p1")
				playerCards = make([]interface{}, len(cardStrings))
				for i, card := range cardStrings {
					playerCards[i] = card
				}
			} else {
				cardStrings := entry.GetStringSlice("p2")
				playerCards = make([]interface{}, len(cardStrings))
				for i, card := range cardStrings {
					playerCards[i] = card
				}
			}

			for _, cardInterface := range playerCards {
				if cardID, ok := cardInterface.(string); ok {
					cardsUsed = append(cardsUsed, cardID)
				}
			}
		}
	}

	// Find the next card in the shuffled deck that hasn't been used
	for _, cardID := range cards {
		used := false
		for _, usedCard := range cardsUsed {
			if cardID == usedCard {
				used = true
				break
			}
		}
		if !used {
			return cardID, nil
		}
	}

	return "", fmt.Errorf("no more cards available in deck for player %d", player)
}

// expandDeckToCards converts a deck's card count map to a list of individual card IDs
func expandDeckToCards(deckCards map[string]int) []string {
	var cards []string
	for cardID, count := range deckCards {
		for i := 0; i < count; i++ {
			cards = append(cards, cardID)
		}
	}
	return cards
}

// TruncateGame truncates a game's log to the provided content
func TruncateGame(gameName, newLogContent string) error {
	db := database.GetDB()
	_, err := db.Exec("UPDATE games SET log_content = $1, modified_at = NOW() WHERE name = $2",
		newLogContent, gameName)
	if err != nil {
		return fmt.Errorf("failed to update game log: %w", err)
	}
	return nil
}

// LoadGameByID is an alias for LoadGame for consistency
func LoadGameByID(gameID string) (*Game, error) {
	id := 0
	if _, err := fmt.Sscanf(gameID, "%d", &id); err != nil {
		return nil, fmt.Errorf("invalid game ID: %s", gameID)
	}
	return LoadGame(id)
}

// TruncateGameByID truncates a game log by ID
func TruncateGameByID(gameID, newLogContent string) error {
	id := 0
	if _, err := fmt.Sscanf(gameID, "%d", &id); err != nil {
		return fmt.Errorf("invalid game ID: %s", gameID)
	}

	db := database.GetDB()
	_, err := db.Exec("UPDATE games SET log_content = $1, modified_at = NOW() WHERE id = $2",
		newLogContent, id)
	if err != nil {
		return fmt.Errorf("failed to update game log: %w", err)
	}
	return nil
}

// AppendActionToGameByID appends an action to a game log by ID
func AppendActionToGameByID(gameID, actionType string, parameters map[string]interface{}) error {
	// Load the game
	gameData, err := LoadGameByID(gameID)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	// Parse current log to get the current state
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		return fmt.Errorf("failed to parse game log: %w", err)
	}

	// Use lens to get current game state instead of manual parsing
	processor := lens.New()
	gameStateData, err := processor.Lens("gameState", entries)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}
	
	gameState := gameStateData.(map[string]interface{})
	currentPlayer := gameState["currentPlayer"].(int)

	// Create the new log entry in event-sourcing format
	eventName := mapActionToEventName(actionType)
	if currentPlayer > 0 {
		parameters["player"] = currentPlayer
	}
	logLine := writeEventLog(eventName, parameters)

	// Start building the updated log content
	newLogContent := gameData.LogContent
	if newLogContent != "" {
		newLogContent += "\n"
	}
	newLogContent += logLine

	// If this is a pass action, we need to add automatic follow-up actions
	if actionType == "pass" {
		currentTurn := gameState["currentTurn"].(int)
		
		// Get the next player (who will start their turn)
		nextPlayer := currentPlayer
		if currentPlayer == 1 {
			nextPlayer = 2
		} else {
			nextPlayer = 1
		}

		// Determine the next turn number
		nextTurn := currentTurn + 1

		// Player 1 doesn't draw on turn 1 (their very first turn), but draws on all other turns
		// Player 2 always draws at the start of their turns
		shouldDraw := !(nextTurn == 1 && nextPlayer == 1)

		// Add draw_card action if the player should draw
		if shouldDraw {
			// Get next card from deck for the player
			nextCard, err := getNextCardFromDeck(entries, gameData, nextPlayer)
			if err == nil && nextCard != "" {
				drawLogLine := writeEventLog("CardDrawn", map[string]interface{}{
					"card_id": nextCard,
					"player":  nextPlayer,
				})
				newLogContent += "\n" + drawLogLine
			}
		}

		// Add turn_start action for the next player
		turnStartLogLine := writeEventLog("TurnStarted", map[string]interface{}{
			"player": nextPlayer,
			"turn":   nextTurn,
		})
		newLogContent += "\n" + turnStartLogLine
	}

	// Update the game in database
	id := 0
	if _, err := fmt.Sscanf(gameID, "%d", &id); err != nil {
		return fmt.Errorf("invalid game ID: %s", gameID)
	}

	db := database.GetDB()
	_, err = db.Exec("UPDATE games SET log_content = $1, modified_at = NOW() WHERE id = $2",
		newLogContent, id)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}
