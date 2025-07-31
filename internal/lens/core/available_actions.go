package core

import (
	"fmt"
	"quards/internal/lens/services"
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

// AvailableActionsLens generates available actions for the current player (pure function)
func AvailableActionsLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	if len(entries) == 0 {
		return []Action{}
	}

	// Get current game state using lenses
	zonesData := ZonesLens(entries, services)
	statsData := PlayerStatsLens(entries, services)

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
	playerKey := fmt.Sprintf("player%d", currentPlayer)
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

	// Hand cards - handle the correct type from zones lens
	handCardsInterface := playerZones["hand"]
	var handCards []HandCard
	if hc, ok := handCardsInterface.([]HandCard); ok {
		handCards = hc
	} else if handInterface, ok := handCardsInterface.([]interface{}); ok {
		// Handle the converted interface{} format from zones lens
		handCards = make([]HandCard, len(handInterface))
		for i, cardInterface := range handInterface {
			if cardMap, ok := cardInterface.(map[string]interface{}); ok {
				if cardID, ok := cardMap["card_id"].(string); ok {
					handCards[i] = HandCard{CardID: cardID}
				}
			}
		}
	}

	// Play card actions
	for _, card := range handCards {
		cardId := card.CardID
		cardInfo := getCardInfo(cardId, services.CardDB)

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
		} else {
			// Card not found in database - still show play action but mark as invalid
			action := Action{
				Type:        "play_card",
				Description: fmt.Sprintf("Play %s", cardId),
				Parameters: map[string]interface{}{
					"card_id": cardId,
					"cost":    0,
				},
				Valid:  false,
				Reason: "Card not found in database",
			}
			actions = append(actions, action)
		}
	}

	// Ink card actions
	for _, card := range handCards {
		cardId := card.CardID
		cardInfo := getCardInfo(cardId, services.CardDB)

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
		} else {
			// Card not found in database - still show ink action but mark as invalid
			action := Action{
				Type:        "ink_card",
				Description: fmt.Sprintf("Ink %s", cardId),
				Parameters: map[string]interface{}{
					"card_id": cardId,
				},
				Valid:  false,
				Reason: "Card not found in database",
			}
			actions = append(actions, action)
		}
	}

	// Quest actions - handle the correct type from zones lens
	playCardsInterface := playerZones["in_play"]
	var playCards []InPlayCard
	if pc, ok := playCardsInterface.([]InPlayCard); ok {
		playCards = pc
	} else if inPlayInterface, ok := playCardsInterface.([]interface{}); ok {
		// Handle the converted interface{} format from zones lens
		playCards = make([]InPlayCard, len(inPlayInterface))
		for i, cardInterface := range inPlayInterface {
			if cardMap, ok := cardInterface.(map[string]interface{}); ok {
				card := InPlayCard{}
				if cardID, ok := cardMap["card_id"].(string); ok {
					card.CardID = cardID
				}
				if instanceID, ok := cardMap["instance_id"].(string); ok {
					card.InstanceID = instanceID
				}
				if exhausted, ok := cardMap["exhausted"].(bool); ok {
					card.Exhausted = exhausted
				}
				if turnPlayed, ok := cardMap["turn_played"].(int); ok {
					card.TurnPlayed = turnPlayed
				} else if turnPlayedFloat, ok := cardMap["turn_played"].(float64); ok {
					card.TurnPlayed = int(turnPlayedFloat)
				}
				playCards[i] = card
			}
		}
	}
	
	for _, card := range playCards {
		cardId := card.CardID
		exhausted := card.Exhausted
		cardInfo := getCardInfo(cardId, services.CardDB)

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
			turnPlayed := card.TurnPlayed
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

	// Convert to interface{} slice for JSON compatibility
	result := make([]interface{}, len(actions))
	for i, action := range actions {
		result[i] = map[string]interface{}{
			"type":        action.Type,
			"description": action.Description,
			"parameters":  action.Parameters,
			"valid":       action.Valid,
			"reason":      action.Reason,
		}
	}
	return result
}

// getCurrentPlayer determines which player should act next
func getCurrentPlayer(entries []parser.LogEntry) int {
	// Look for the most recent turn_start or pass action
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Event == parser.TurnStarted {
			return entry.GetPlayer()
		} else if entry.Event == parser.TurnPassed {
			// After a pass, it's the other player's turn
			player := entry.GetPlayer()
			if player == 1 {
				return 2
			} else if player == 2 {
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
		if entry.Event == parser.TurnStarted {
			currentTurn = entry.GetInt("turn")
		}
	}
	return currentTurn
}


// getCardInfo retrieves card information from the database
func getCardInfo(cardId string, cardDB services.CardDatabase) map[string]interface{} {
	if cardData, exists := cardDB.GetCard(cardId); exists {
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