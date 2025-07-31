package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// PlayerStats represents computed player statistics
type PlayerStats struct {
	Player1 PlayerStatValues `json:"player1"`
	Player2 PlayerStatValues `json:"player2"`
}

type PlayerStatValues struct {
	Lore              int `json:"lore"`
	TotalInk          int `json:"total_ink"`
	AvailableInk      int `json:"available_ink"`
	InksThisTurn      int `json:"inks_this_turn"`
	AvailableInkings  int `json:"available_inkings"`
}

// InkwellCard represents a card in the inkwell
type InkwellCard struct {
	CardID string `json:"cardId"`
}

// PlayerStatsLens computes player statistics from zones data and game log (pure function)
func PlayerStatsLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	// Get zones data first - this is the foundational lens
	zonesData := ZonesLens(entries, services)
	zones, ok := zonesData.(map[string]interface{})
	if !ok {
		// Fallback if zones data is malformed
		return createEmptyPlayerStats()
	}

	stats := PlayerStats{
		Player1: PlayerStatValues{Lore: 0, TotalInk: 0, AvailableInk: 0, InksThisTurn: 0, AvailableInkings: 1},
		Player2: PlayerStatValues{Lore: 0, TotalInk: 0, AvailableInk: 0, InksThisTurn: 0, AvailableInkings: 1},
	}

	// Extract inkwell data from zones instead of re-processing
	var player1Inkwell, player2Inkwell []InkwellCard
	if p1Data, ok := zones["player1"].(map[string]interface{}); ok {
		if inkData, ok := p1Data["ink"].([]InkCard); ok {
			for _, ink := range inkData {
				player1Inkwell = append(player1Inkwell, InkwellCard{CardID: ink.CardID})
			}
		}
	}
	if p2Data, ok := zones["player2"].(map[string]interface{}); ok {
		if inkData, ok := p2Data["ink"].([]InkCard); ok {
			for _, ink := range inkData {
				player2Inkwell = append(player2Inkwell, InkwellCard{CardID: ink.CardID})
			}
		}
	}

	cardDB := services.CardDB.GetAll()
	currentTurn := 0

	for _, entry := range entries {
		switch entry.Event {
		case parser.QuestAttempted:
			// Quest actions gain lore
			_ = entry.GetInstance("instance") // TODO: Use instance ID for tracking
			loreValue := entry.GetInt("lore")
			if loreValue == 0 {
				loreValue = 1 // Default fallback
			}
			player := entry.GetPlayer()
			
			if player == 1 {
				stats.Player1.Lore += loreValue
			} else if player == 2 {
				stats.Player2.Lore += loreValue
			}

		case parser.CardInked:
			// Ink actions increase total ink AND available ink (immediately usable)
			// Also decreases available inkings for this turn
			// Note: Inkwell contents come from zones data, only track stats here
			player := entry.GetPlayer()
			if player == 1 {
				stats.Player1.TotalInk++
				stats.Player1.AvailableInk++  // Newly inked card is immediately available
				stats.Player1.InksThisTurn++
				stats.Player1.AvailableInkings-- // Can't ink another card this turn
			} else if player == 2 {
				stats.Player2.TotalInk++
				stats.Player2.AvailableInk++  // Newly inked card is immediately available
				stats.Player2.InksThisTurn++
				stats.Player2.AvailableInkings-- // Can't ink another card this turn
			}

		case parser.CardPlayed:
			// Playing cards spends ink - get cost from card database
			cardID := entry.GetCard("card_id")
			player := entry.GetPlayer()
			costInt := 0
			
			if card, exists := cardDB[cardID]; exists {
				costInt = card.Cost
			}

			if player == 1 {
				stats.Player1.AvailableInk -= costInt
				if stats.Player1.AvailableInk < 0 {
					stats.Player1.AvailableInk = 0
				}
			} else if player == 2 {
				stats.Player2.AvailableInk -= costInt
				if stats.Player2.AvailableInk < 0 {
					stats.Player2.AvailableInk = 0
				}
			}

		case parser.TurnPassed:
			// On any pass, reset all turn-based stats to defaults
			stats.Player1.AvailableInk = stats.Player1.TotalInk
			stats.Player2.AvailableInk = stats.Player2.TotalInk
			stats.Player1.InksThisTurn = 0
			stats.Player1.AvailableInkings = 1
			stats.Player2.InksThisTurn = 0
			stats.Player2.AvailableInkings = 1

		case parser.TurnStarted:
			// At turn start, just track the turn number
			currentTurn++
		}
	}

	// Convert to interface{} format for both available actions lens and frontend
	result := make(map[string]interface{})

	// Player 1 data
	player1Data := map[string]interface{}{
		"lore":               stats.Player1.Lore,
		"total_ink":          stats.Player1.TotalInk,
		"available_ink":      stats.Player1.AvailableInk,
		"inks_this_turn":     stats.Player1.InksThisTurn,
		"available_inkings":  stats.Player1.AvailableInkings,
		"inkwell":            player1Inkwell,
	}
	result["player1"] = player1Data

	// Player 2 data
	player2Data := map[string]interface{}{
		"lore":               stats.Player2.Lore,
		"total_ink":          stats.Player2.TotalInk,
		"available_ink":      stats.Player2.AvailableInk,
		"inks_this_turn":     stats.Player2.InksThisTurn,
		"available_inkings":  stats.Player2.AvailableInkings,
		"inkwell":            player2Inkwell,
	}
	result["player2"] = player2Data

	return result
}

// createEmptyPlayerStats returns empty player stats as fallback
func createEmptyPlayerStats() interface{} {
	result := make(map[string]interface{})

	emptyData := map[string]interface{}{
		"lore":               0,
		"total_ink":          0,
		"available_ink":      0,
		"inks_this_turn":     0,
		"available_inkings":  1,
		"inkwell":            []InkwellCard{},
	}

	result["player1"] = emptyData
	result["player2"] = emptyData

	return result
}