package game

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
	
	"quards/internal/database"
	"quards/internal/deck"
	"quards/internal/parser"
)

// Game represents a game instance
type Game struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"` // Required
	Player1Deck  string    `json:"player1Deck"`
	Player2Deck  string    `json:"player2Deck"`
	Seed         *int      `json:"seed"`
	LogContent   string    `json:"logContent"`
	Status       string    `json:"status"`
	Winner       *int      `json:"winner"`
	Turns        int       `json:"turns"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// GameList represents a simplified game list for API responses
type GameList struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"` // Required
	Player1Deck  string    `json:"player1Deck"`
	Player2Deck  string    `json:"player2Deck"`
	Seed         *int      `json:"seed"`
	Status       string    `json:"status"`
	Winner       *int      `json:"winner"`
	Turns        int       `json:"turns"`
	Created      time.Time `json:"created"`
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	Name        string `json:"name,omitempty"` // Optional, will auto-generate if empty
	Player1Deck string `json:"player1Deck"`
	Player2Deck string `json:"player2Deck"`
	Seed        *int   `json:"seed"`
	LogContent  string `json:"logContent,omitempty"` // For uploaded games
}

// CreateGame creates a new game and generates initial log
func CreateGame(req *CreateGameRequest) (*Game, error) {
	db := database.GetDB()
	
	// Generate seed if not provided
	seed := req.Seed
	if seed == nil {
		generatedSeed := rand.Intn(1000000)
		seed = &generatedSeed
	}
	
	// Generate game name if not provided
	gameName := req.Name
	if gameName == "" {
		gameName = fmt.Sprintf("game-%d", time.Now().Unix())
	}
	
	// Use provided log content or generate initial log
	var logContent string
	if req.LogContent != "" {
		logContent = req.LogContent
	} else {
		logContent = generateInitialLog(req.Player1Deck, req.Player2Deck, *seed)
	}
	
	// Insert game into database with required name
	var gameID int
	err := db.QueryRow(`
		INSERT INTO games (name, player1_deck, player2_deck, seed, log_content, status)
		VALUES ($1, $2, $3, $4, $5, 'created')
		RETURNING id`,
		gameName, req.Player1Deck, req.Player2Deck, seed, logContent).Scan(&gameID)
	
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
		SELECT id, name, player1_deck, player2_deck, seed, log_content, 
		       status, winner, turns, created_at, modified_at
		FROM games WHERE id = $1`, id).Scan(
		&game.ID, &game.Name, &game.Player1Deck, &game.Player2Deck,
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
		SELECT id, name, player1_deck, player2_deck, seed, log_content,
		       status, winner, turns, created_at, modified_at
		FROM games WHERE name = $1`, name).Scan(
		&game.ID, &game.Name, &game.Player1Deck, &game.Player2Deck,
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
		SELECT id, name, player1_deck, player2_deck, seed, status, 
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
		err := rows.Scan(&game.ID, &game.Name, &game.Player1Deck, &game.Player2Deck,
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
		// Fallback to test cards if deck loading fails
		return generateFallbackLog(player1DeckName, player2DeckName, seed)
	}
	
	player2DeckData, err := deck.LoadDeck(player2DeckName)
	if err != nil {
		// Fallback to test cards if deck loading fails
		return generateFallbackLog(player1DeckName, player2DeckName, seed)
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
	startEntry := parser.CreateLogEntry(0, 0, "game_start", map[string]interface{}{
		"player1_deck": player1DeckName,
		"player2_deck": player2DeckName,
		"seed":         seed,
	})
	startData, _ := json.Marshal(startEntry)
	entries = append(entries, string(startData))
	
	// Shuffle decks
	shuffleEntry := parser.CreateLogEntry(0, 0, "shuffle_decks", map[string]interface{}{
		"seed": seed,
	})
	shuffleData, _ := json.Marshal(shuffleEntry)
	entries = append(entries, string(shuffleData))
	
	// Draw opening hands - single action with all cards for both players
	openingHandsEntry := parser.CreateLogEntry(0, 0, "draw_opening_hands", map[string]interface{}{
		"player1_cards": player1Cards[:7], // First 7 cards for player 1
		"player2_cards": player2Cards[:7], // First 7 cards for player 2
	})
	openingHandsData, _ := json.Marshal(openingHandsEntry)
	entries = append(entries, string(openingHandsData))
	
	// Player 1 starts (turn 1)
	turnEntry := parser.CreateLogEntry(1, 1, "turn_start", map[string]interface{}{
		"player": 1,
		"turn":   1,
	})
	turnData, _ := json.Marshal(turnEntry)
	entries = append(entries, string(turnData))
	
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

	// Determine the current player from the most recent turn_start or pass action
	currentPlayer := 1 // Default to player 1
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Action == "turn_start" && entry.Parameters != nil {
			if player, ok := entry.Parameters["player"]; ok {
				switch v := player.(type) {
				case float64:
					currentPlayer = int(v)
				case int:
					currentPlayer = v
				}
				break
			}
		} else if entry.Action == "pass" {
			// After a pass, it's the other player's turn
			if entry.Player == 1 {
				currentPlayer = 2
			} else if entry.Player == 2 {
				currentPlayer = 1
			}
			break
		}
	}

	// Get current turn number
	currentTurn := 1
	for _, entry := range entries {
		if entry.Action == "turn_start" && entry.Parameters != nil {
			if turn, ok := entry.Parameters["turn"]; ok {
				switch v := turn.(type) {
				case float64:
					currentTurn = int(v)
				case int:
					currentTurn = v
				}
			}
		}
	}

	// Create the new log entry
	newEntry := parser.CreateLogEntry(currentTurn, currentPlayer, actionType, parameters)
	newEntryData, err := json.Marshal(newEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal new entry: %w", err)
	}

	// Start building the updated log content
	updatedLogContent := gameData.LogContent + string(newEntryData) + "\n"

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
		shouldDraw := !(nextPlayer == 1 && nextTurn == 1)

		// Add draw_card action if the player should draw
		if shouldDraw {
			// Get next card from deck for the player
			nextCard, err := getNextCardFromDeck(entries, gameData, nextPlayer)
			if err == nil && nextCard != "" {
				drawEntry := parser.CreateLogEntry(nextTurn, nextPlayer, "draw_card", map[string]interface{}{
					"card_id": nextCard,
				})
				drawEntryData, err := json.Marshal(drawEntry)
				if err == nil {
					updatedLogContent += string(drawEntryData) + "\n"
				}
			}
		}

		// Add turn_start action for the next player
		turnStartEntry := parser.CreateLogEntry(nextTurn, nextPlayer, "turn_start", map[string]interface{}{
			"player": nextPlayer,
			"turn":   nextTurn,
		})
		turnStartEntryData, err := json.Marshal(turnStartEntry)
		if err == nil {
			updatedLogContent += string(turnStartEntryData) + "\n"
		}
	}

	// Update the game in the database
	db := database.GetDB()
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
		if entry.Action == "game_start" && entry.Parameters != nil {
			if deck1, ok := entry.Parameters["player1_deck"].(string); ok {
				player1DeckName = deck1
			}
			if deck2, ok := entry.Parameters["player2_deck"].(string); ok {
				player2DeckName = deck2
			}
			if seedFloat, ok := entry.Parameters["seed"].(float64); ok {
				seed = int(seedFloat)
			} else if seedInt, ok := entry.Parameters["seed"].(int); ok {
				seed = seedInt
			}
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
		if entry.Player == player {
			switch entry.Action {
			case "draw_card", "ink_card", "play_card":
				if entry.Parameters != nil {
					if cardID, ok := entry.Parameters["card_id"].(string); ok {
						cardsUsed = append(cardsUsed, cardID)
					}
				}
			}
		} else if entry.Action == "draw_opening_hands" && entry.Parameters != nil {
			// Handle opening hands
			var playerCards []interface{}
			if player == 1 {
				if cards, ok := entry.Parameters["player1_cards"].([]interface{}); ok {
					playerCards = cards
				}
			} else {
				if cards, ok := entry.Parameters["player2_cards"].([]interface{}); ok {
					playerCards = cards
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

// generateFallbackLog creates a fallback log with test cards if deck loading fails
func generateFallbackLog(player1DeckName, player2DeckName string, seed int) string {
	var entries []string
	
	// Game start entry
	startEntry := parser.CreateLogEntry(0, 0, "game_start", map[string]interface{}{
		"player1_deck": player1DeckName,
		"player2_deck": player2DeckName,
		"seed":         seed,
	})
	startData, _ := json.Marshal(startEntry)
	entries = append(entries, string(startData))
	
	// Shuffle decks
	shuffleEntry := parser.CreateLogEntry(0, 0, "shuffle_decks", map[string]interface{}{
		"seed": seed,
	})
	shuffleData, _ := json.Marshal(shuffleEntry)
	entries = append(entries, string(shuffleData))
	
	// Fallback to test cards
	player1Cards := []string{"TFC-001", "TFC-002", "TFC-003", "TFC-004", "TFC-005", "TFC-006", "TFC-007"}
	player2Cards := []string{"TFC-008", "TFC-009", "TFC-010", "TFC-011", "TFC-012", "TFC-013", "TFC-014"}
	
	// Draw opening hands - single action with all cards for both players
	openingHandsEntry := parser.CreateLogEntry(0, 0, "draw_opening_hands", map[string]interface{}{
		"player1_cards": player1Cards,
		"player2_cards": player2Cards,
	})
	openingHandsData, _ := json.Marshal(openingHandsEntry)
	entries = append(entries, string(openingHandsData))
	
	// Player 1 starts (turn 1)
	turnEntry := parser.CreateLogEntry(1, 1, "turn_start", map[string]interface{}{
		"player": 1,
		"turn":   1,
	})
	turnData, _ := json.Marshal(turnEntry)
	entries = append(entries, string(turnData))
	
	return strings.Join(entries, "\n") + "\n"
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
