package lens

import (
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
	InksThisTurn      int `json:"inks_this_turn"`      // How many cards inked this turn
	AvailableInkings  int `json:"available_inkings"`   // How many more cards can be inked this turn (default 1)
}

// InkwellCard represents a card in the inkwell
type InkwellCard struct {
	CardID string `json:"cardId"` // The card ID
}

// PlayerStatsLens computes player statistics from zones data and game log
func PlayerStatsLens(entries []parser.LogEntry) interface{} {
	// Get zones data first - this is the foundational lens
	zonesData := ZonesLens(entries)
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
	
	cardDB := GetCardDatabase()
	currentTurn := 0
	
	for _, entry := range entries {
		switch entry.Action {
		case "quest":
			// Quest actions gain lore
			if entry.Parameters != nil {
				if cardID, ok := entry.Parameters["card_id"].(string); ok {
					// Look up the card's lore value
					loreValue := 1 // Default fallback
					if card, exists := cardDB[cardID]; exists {
						loreValue = card.Lore
					}
					
					if entry.Player == 1 {
						stats.Player1.Lore += loreValue
					} else if entry.Player == 2 {
						stats.Player2.Lore += loreValue
					}
				}
			}
			
		case "ink_card":
			// Ink actions increase total ink AND available ink (immediately usable)
			// Also decreases available inkings for this turn
			// Note: Inkwell contents come from zones data, only track stats here
			if entry.Player == 1 {
				stats.Player1.TotalInk++
				stats.Player1.AvailableInk++  // Newly inked card is immediately available
				stats.Player1.InksThisTurn++
				stats.Player1.AvailableInkings-- // Can't ink another card this turn
			} else if entry.Player == 2 {
				stats.Player2.TotalInk++
				stats.Player2.AvailableInk++  // Newly inked card is immediately available
				stats.Player2.InksThisTurn++
				stats.Player2.AvailableInkings-- // Can't ink another card this turn
			}
			
		case "play_card":
			// Playing cards spends ink
			if entry.Parameters != nil {
				if cost, ok := entry.Parameters["cost"]; ok {
					costInt := 0
					if costFloat, ok := cost.(float64); ok {
						costInt = int(costFloat)
					} else if costIntVal, ok := cost.(int); ok {
						costInt = costIntVal
					}
					
					if entry.Player == 1 {
						stats.Player1.AvailableInk -= costInt
						if stats.Player1.AvailableInk < 0 {
							stats.Player1.AvailableInk = 0
						}
					} else if entry.Player == 2 {
						stats.Player2.AvailableInk -= costInt
						if stats.Player2.AvailableInk < 0 {
							stats.Player2.AvailableInk = 0
						}
					}
				}
			}
			
		case "pass":
			// On any pass, reset all turn-based stats to defaults
			stats.Player1.AvailableInk = stats.Player1.TotalInk
			stats.Player2.AvailableInk = stats.Player2.TotalInk
			stats.Player1.InksThisTurn = 0
			stats.Player1.AvailableInkings = 1
			stats.Player2.InksThisTurn = 0
			stats.Player2.AvailableInkings = 1
			
		case "turn_start":
			// At turn start, just track the turn number
			currentTurn++
		}
	}
	
	// Convert to interface{} format for both available actions lens and frontend
	result := make(map[string]interface{})
	
	// Player 1 data (both formats for compatibility)
	player1Data := map[string]interface{}{
		"lore":               stats.Player1.Lore,
		"total_ink":          stats.Player1.TotalInk,
		"available_ink":      stats.Player1.AvailableInk,
		"inks_this_turn":     stats.Player1.InksThisTurn,
		"available_inkings":  stats.Player1.AvailableInkings,
		"inkwell":            player1Inkwell,
	}
	result["player_1"] = player1Data // For available actions lens
	result["player1"] = player1Data  // For frontend compatibility
	
	// Player 2 data (both formats for compatibility)
	player2Data := map[string]interface{}{
		"lore":               stats.Player2.Lore,
		"total_ink":          stats.Player2.TotalInk,
		"available_ink":      stats.Player2.AvailableInk,
		"inks_this_turn":     stats.Player2.InksThisTurn,
		"available_inkings":  stats.Player2.AvailableInkings,
		"inkwell":            player2Inkwell,
	}
	result["player_2"] = player2Data // For available actions lens
	result["player2"] = player2Data  // For frontend compatibility
	
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
	
	result["player_1"] = emptyData
	result["player1"] = emptyData
	result["player_2"] = emptyData
	result["player2"] = emptyData
	
	return result
}