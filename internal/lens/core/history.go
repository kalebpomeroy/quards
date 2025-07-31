package core

import (
	"fmt"
	"quards/internal/lens/services"
	"quards/internal/parser"
	"strings"
)

// HistoryEntry represents a human-readable game event
type HistoryEntry struct {
	Step        int               `json:"step"`
	Player      int               `json:"player"`
	Event       string            `json:"event"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
}

// HistoryLens converts raw log events into human-readable history entries (pure function)
func HistoryLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	history := make([]HistoryEntry, len(entries))

	for i, entry := range entries {
		historyEntry := HistoryEntry{
			Step:        entry.Step,
			Player:      entry.GetPlayer(),
			Event:       string(entry.Event),
			Parameters:  entry.Parameters,
			Description: formatEventDescription(entry, services.CardDB),
		}

		history[i] = historyEntry
	}

	return history
}

// formatEventDescription converts a log entry into human-readable text
func formatEventDescription(entry parser.LogEntry, cardDB services.CardDatabase) string {
	player := entry.GetPlayer()
	playerStr := ""
	if player > 0 {
		playerStr = fmt.Sprintf("Player %d ", player)
	}

	switch entry.Event {
	case parser.GameStarted:
		p1Deck := entry.Parameters["p1_deck"]
		p2Deck := entry.Parameters["p2_deck"]
		seed := entry.Parameters["seed"]
		return fmt.Sprintf("Game starts: %s vs %s (Seed: %s)", p1Deck, p2Deck, seed)

	case parser.DecksShuffled:
		seed := entry.Parameters["seed"]
		return fmt.Sprintf("Decks shuffled (Seed: %s)", seed)

	case parser.OpeningHandsDrawn:
		return "Players draw opening hands (7 cards each)"

	case parser.TurnStarted:
		turn := entry.GetInt("turn")
		return fmt.Sprintf("%sstarts turn %d", playerStr, turn)

	case parser.CardDrawn:
		cardID := entry.GetCard("card")
		cardName := getCardName(cardID, cardDB)
		return fmt.Sprintf("%sdraws %s", playerStr, cardName)

	case parser.CardInked:
		cardID := entry.GetCard("card_id")
		cardName := getCardName(cardID, cardDB)
		return fmt.Sprintf("%sinks %s", playerStr, cardName)

	case parser.CardPlayed:
		cardID := entry.GetCard("card_id")
		cardName := getCardName(cardID, cardDB)
		cost := entry.GetInt("cost")
		costStr := ""
		if cost > 0 {
			costStr = fmt.Sprintf(" (%d ink)", cost)
		}
		return fmt.Sprintf("%splays %s%s", playerStr, cardName, costStr)

	case parser.QuestAttempted:
		cardID := entry.GetCard("card_id")
		if cardID == "" {
			// Try instance-based lookup
			instanceID := entry.GetInstance("instance")
			cardName := fmt.Sprintf("Character %s", instanceID)
			if cardID != "" {
				cardName = getCardName(cardID, cardDB)
			}
			lore := entry.GetInt("lore")
			loreStr := ""
			if lore > 0 {
				loreStr = fmt.Sprintf(" (+%d lore)", lore)
			}
			return fmt.Sprintf("%squests with %s%s", playerStr, cardName, loreStr)
		}
		cardName := getCardName(cardID, cardDB)
		lore := entry.GetInt("lore")
		loreStr := ""
		if lore > 0 {
			loreStr = fmt.Sprintf(" (+%d lore)", lore)
		}
		return fmt.Sprintf("%squests with %s%s", playerStr, cardName, loreStr)

	case parser.TurnPassed:
		return fmt.Sprintf("%spasses turn", playerStr)

	case parser.CharacterExerted:
		instanceID := entry.GetInstance("instance")
		return fmt.Sprintf("Character %s becomes exhausted", instanceID)

	case parser.CharacterReadied:
		instanceID := entry.GetInstance("instance")
		return fmt.Sprintf("Character %s becomes ready", instanceID)

	case parser.CharacterBanished:
		instanceID := entry.GetInstance("instance")
		return fmt.Sprintf("Character %s is banished", instanceID)

	default:
		// Generic fallback for unknown events
		paramStr := ""
		if entry.Parameters != nil && len(entry.Parameters) > 0 {
			var params []string
			for k, v := range entry.Parameters {
				params = append(params, fmt.Sprintf("%s: %v", k, v))
			}
			paramStr = fmt.Sprintf(" (%s)", strings.Join(params, ", "))
		}
		return fmt.Sprintf("%s%s%s", playerStr, strings.ToLower(string(entry.Event)), paramStr)
	}
}

// getCardName returns the card name from the database, or the card ID as fallback
func getCardName(cardID string, cardDB services.CardDatabase) string {
	if cardID == "" {
		return "unknown card"
	}

	if cardData, exists := cardDB.GetCard(cardID); exists {
		return cardData.Name
	}

	return cardID // Fallback to ID if not found in database
}