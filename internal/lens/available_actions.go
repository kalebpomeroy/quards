package lens

import (
	"fmt"
	"quards/internal/parser"
)

// Action represents a possible action a player can take
type Action struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Valid       bool                   `json:"valid"`
	Reason      string                 `json:"reason,omitempty"` // Why action is invalid
}

// AvailableActionsLens generates available actions for the current player
func AvailableActionsLens(entries []parser.LogEntry) interface{} {
	if len(entries) == 0 {
		return []Action{}
	}

	// Get current game state using existing lenses
	zonesData := ZonesLens(entries)
	statsData := PlayerStatsLens(entries)
	
	zones := zonesData.(map[string]interface{})
	stats := statsData.(map[string]interface{})
	
	// Determine current player from last action or game state
	currentPlayer := getCurrentPlayer(entries)
	if currentPlayer == 0 {
		// Game hasn't started or is system turn
		return []Action{}
	}
	
	// Determine current turn number
	currentTurn := getCurrentTurn(entries)
	
	var actions []Action
	
	// Always available: Pass
	actions = append(actions, Action{
		Type:        "pass",
		Description: "End turn and pass to opponent",
		Parameters:  map[string]interface{}{},
		Valid:       true,
	})
	
	// Get player-specific data
	playerKey := getPlayerKey(currentPlayer)
	playerZones := zones[playerKey].(map[string]interface{})
	playerStats := stats[playerKey].(map[string]interface{})
	
	// Available ink - handle both int and float64
	var availableInk int
	switch v := playerStats["available_ink"].(type) {
	case float64:
		availableInk = int(v)
	case int:
		availableInk = v
	default:
		availableInk = 0
	}
	
	// Available inkings this turn
	var availableInkings int
	switch v := playerStats["available_inkings"].(type) {
	case float64:
		availableInkings = int(v)
	case int:
		availableInkings = v
	default:
		availableInkings = 1 // Default to 1 if not found
	}
	
	// Hand cards
	handCards := playerZones["hand"].([]interface{})
	
	// Play card actions
	for _, cardInterface := range handCards {
		cardData := cardInterface.(map[string]interface{})
		cardId := cardData["card_id"].(string)
		cardInfo := getCardInfo(cardId)
		
		if cardInfo != nil {
			var cost int
			switch v := cardInfo["Cost"].(type) {
			case float64:
				cost = int(v)
			case int:
				cost = v
			default:
				cost = 0
			}
			canPlay := cost <= availableInk
			
			action := Action{
				Type:        "play_card",
				Description: fmt.Sprintf("Play %s (Cost: %d)", cardInfo["Name"], cost),
				Parameters: map[string]interface{}{
					"card_id": cardId,
					"cost":    cost,
				},
				Valid: canPlay,
			}
			
			if !canPlay {
				action.Reason = fmt.Sprintf("Not enough ink (need %d, have %d)", cost, availableInk)
			}
			
			actions = append(actions, action)
		}
	}
	
	// Ink card actions
	for _, cardInterface := range handCards {
		cardData := cardInterface.(map[string]interface{})
		cardId := cardData["card_id"].(string)
		cardInfo := getCardInfo(cardId)
		
		if cardInfo != nil {
			inkable := cardInfo["Inkable"].(bool)
			canInkThisTurn := availableInkings > 0
			
			valid := inkable && canInkThisTurn
			
			action := Action{
				Type:        "ink_card",
				Description: fmt.Sprintf("Ink %s", cardInfo["Name"]),
				Parameters: map[string]interface{}{
					"card_id": cardId,
				},
				Valid: valid,
			}
			
			if !inkable {
				action.Reason = "Card is not inkable"
			} else if !canInkThisTurn {
				action.Reason = "Already inked a card this turn"
			}
			
			actions = append(actions, action)
		}
	}
	
	// Quest actions
	playCards := playerZones["in_play"].([]interface{})
	for _, cardInterface := range playCards {
		cardData := cardInterface.(map[string]interface{})
		cardId := cardData["card_id"].(string)
		exhausted := cardData["exhausted"].(bool)
		cardInfo := getCardInfo(cardId)
		
		if cardInfo != nil {
			cardType := cardInfo["Type"].(string)
			lore := cardInfo["Lore"]
			
			// Handle lore value safely  
			var loreValue int
			var hasLore bool
			switch v := lore.(type) {
			case float64:
				loreValue = int(v)
				hasLore = v > 0
			case int:
				loreValue = v
				hasLore = v > 0
			default:
				loreValue = 0
				hasLore = false
			}
			
			// Check if character is wet (sick) - characters are wet the turn they were played
			var turnPlayed int
			if turnPlayedInterface, ok := cardData["turn_played"]; ok {
				switch v := turnPlayedInterface.(type) {
				case float64:
					turnPlayed = int(v)
				case int:
					turnPlayed = v
				default:
					turnPlayed = 0
				}
			}
			
			isWet := turnPlayed == currentTurn
			
			// Only characters with lore can quest, and they must be dry (not played this turn)
			canQuest := cardType == "Character" && hasLore && !exhausted && !isWet
			
			action := Action{
				Type:        "quest",
				Description: fmt.Sprintf("Quest with %s (Lore: %d)", cardInfo["Name"], loreValue),
				Parameters: map[string]interface{}{
					"card_id": cardId,
					"lore":    loreValue,
				},
				Valid: canQuest,
			}
			
			if exhausted {
				action.Reason = "Character is exhausted"
			} else if cardType != "Character" {
				action.Reason = "Only characters can quest"
			} else if !hasLore {
				action.Reason = "Character has no lore value"
			} else if isWet {
				action.Reason = "Character is wet (played this turn)"
			}
			
			actions = append(actions, action)
		}
	}
	
	return actions
}

// getCurrentPlayer determines which player should act next
func getCurrentPlayer(entries []parser.LogEntry) int {
	// Look for the most recent turn_start or pass action
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Action == "turn_start" {
			// Extract player from structured parameters
			if entry.Parameters != nil {
				switch v := entry.Parameters["player"].(type) {
				case float64:
					return int(v)
				case int:
					return v
				}
			}
		} else if entry.Action == "pass" {
			// After a pass, it's the other player's turn
			if entry.Player == 1 {
				return 2
			} else if entry.Player == 2 {
				return 1
			}
		}
	}
	
	// Default to player 1 if no turn_start or pass found
	return 1
}

// getCurrentTurn determines the current turn number from game log
func getCurrentTurn(entries []parser.LogEntry) int {
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
	return currentTurn
}

// getPlayerKey returns the map key for player data
func getPlayerKey(player int) string {
	return fmt.Sprintf("player_%d", player)
}


// getCardInfo retrieves card information from the database
func getCardInfo(cardId string) map[string]interface{} {
	// Use global card database
	cardDB := GetCardDatabase()
	if cardData, exists := cardDB[cardId]; exists {
		// Convert CardData struct to map for easier access
		cardMap := make(map[string]interface{})
		cardMap["Name"] = cardData.Name
		cardMap["Cost"] = float64(cardData.Cost)
		cardMap["Inkable"] = cardData.Inkable
		cardMap["Type"] = cardData.Type
		cardMap["Lore"] = float64(cardData.Lore)
		return cardMap
	}
	return nil
}