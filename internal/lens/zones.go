package lens

import (
	"quards/internal/parser"
	"strconv"
)

// GameZones represents the state of all zones for both players
type GameZones struct {
	Player1 PlayerZones `json:"player1"`
	Player2 PlayerZones `json:"player2"`
}

// PlayerZones represents all zones for a single player
type PlayerZones struct {
	Hand        []string               `json:"hand"`        // Hand card IDs (for available actions)
	Deck        int                    `json:"deck"`        // Hidden count
	Discard     int                    `json:"discard"`     // Hidden count
	DiscardPile []DiscardCard          `json:"discardPile"` // Actual discard pile contents
	Ink         []InkCard              `json:"ink"`         // Ink cards with turn played
	Battlefield []BattlefieldCard      `json:"battlefield"` // Face-up unique cards
}

// DiscardCard represents a card in the discard pile
type DiscardCard struct {
	CardID string `json:"cardId"` // The card ID
}

// InkCard represents a card in the ink zone with turn information
type InkCard struct {
	CardID     string `json:"cardId"`     // The card ID
	TurnPlayed int    `json:"turnPlayed"` // Turn when this was inked
}

// BattlefieldCard represents a unique card instance on the battlefield
type BattlefieldCard struct {
	ID         string `json:"id"`         // Unique instance ID
	CardID     string `json:"cardId"`     // The actual card (ARI-028)
	Exhausted  bool   `json:"exhausted"`  // Whether the card is exhausted (turned 90 degrees)
	TurnPlayed int    `json:"turnPlayed"` // Turn when this card was played (for wet/dry status)
}

var nextCardInstanceID = 1

// ZonesLens tracks card movements between different game zones
func ZonesLens(entries []parser.LogEntry) interface{} {
	zones := GameZones{
		Player1: PlayerZones{
			Hand:        []string{},
			Deck:        60,
			Discard:     0,
			DiscardPile: []DiscardCard{},
			Ink:         []InkCard{},
			Battlefield: []BattlefieldCard{},
		},
		Player2: PlayerZones{
			Hand:        []string{},
			Deck:        60,
			Discard:     0,
			DiscardPile: []DiscardCard{},
			Ink:         []InkCard{},
			Battlefield: []BattlefieldCard{},
		},
	}
	
	currentTurn := 0
	player1FirstTurn := true
	
	// Track battlefield cards for quest operations
	battlefieldCards := make(map[string]*BattlefieldCard)
	
	for _, entry := range entries {
		switch entry.Action {
		case "draw_opening_hands":
			// Handle opening hands drawing for both players at once
			if entry.Parameters != nil {
				if player1Cards, ok := entry.Parameters["player1_cards"].([]interface{}); ok {
					for _, cardInterface := range player1Cards {
						if cardID, ok := cardInterface.(string); ok {
							zones.Player1.Hand = append(zones.Player1.Hand, cardID)
							zones.Player1.Deck--
						}
					}
				}
				if player2Cards, ok := entry.Parameters["player2_cards"].([]interface{}); ok {
					for _, cardInterface := range player2Cards {
						if cardID, ok := cardInterface.(string); ok {
							zones.Player2.Hand = append(zones.Player2.Hand, cardID)
							zones.Player2.Deck--
						}
					}
				}
			}
			
		case "draw_card":
			// Add specific card to player's hand (for individual card draws during gameplay)
			if entry.Parameters != nil {
				if cardID, ok := entry.Parameters["card_id"].(string); ok {
					if entry.Player == 1 {
						zones.Player1.Hand = append(zones.Player1.Hand, cardID)
						zones.Player1.Deck--
					} else if entry.Player == 2 {
						zones.Player2.Hand = append(zones.Player2.Hand, cardID)
						zones.Player2.Deck--
					}
				}
			}
			
		case "ink_card":
			// Move card from hand to ink zone
			if entry.Parameters != nil {
				if cardID, ok := entry.Parameters["card_id"].(string); ok {
					inkCard := InkCard{
						CardID:     cardID,
						TurnPlayed: currentTurn,
					}
					if entry.Player == 1 {
						// Remove card from hand
						zones.Player1.Hand = removeCardFromHand(zones.Player1.Hand, cardID)
						zones.Player1.Ink = append(zones.Player1.Ink, inkCard)
					} else if entry.Player == 2 {
						zones.Player2.Hand = removeCardFromHand(zones.Player2.Hand, cardID)
						zones.Player2.Ink = append(zones.Player2.Ink, inkCard)
					}
				}
			}
			
		case "play_card":
			// Move card from hand to battlefield OR discard (depending on card type)
			if entry.Parameters != nil {
				if cardID, ok := entry.Parameters["card_id"].(string); ok {
					// Use global card database to check card type
					cardDB := GetCardDatabase()
					cardData, exists := cardDB[cardID]
					
					if entry.Player == 1 {
						zones.Player1.Hand = removeCardFromHand(zones.Player1.Hand, cardID)
					} else if entry.Player == 2 {
						zones.Player2.Hand = removeCardFromHand(zones.Player2.Hand, cardID)
					}
					
					// Characters and Items go to battlefield, Actions/Songs go to discard
					if exists && (cardData.Type == "Character" || cardData.Type == "Item") {
						// Move to battlefield
						instanceID := "instance-" + strconv.Itoa(nextCardInstanceID)
						nextCardInstanceID++
						
						battlefieldCard := BattlefieldCard{
							ID:         instanceID,
							CardID:     cardID,
							Exhausted:  false,
							TurnPlayed: currentTurn,
						}
						
						// Track for future quest operations
						battlefieldCards[instanceID] = &battlefieldCard
						
						if entry.Player == 1 {
							zones.Player1.Battlefield = append(zones.Player1.Battlefield, battlefieldCard)
						} else if entry.Player == 2 {
							zones.Player2.Battlefield = append(zones.Player2.Battlefield, battlefieldCard)
						}
					} else {
						// Move to discard pile (Actions, Songs, or unknown cards)
						discardCard := DiscardCard{CardID: cardID}
						if entry.Player == 1 {
							zones.Player1.DiscardPile = append(zones.Player1.DiscardPile, discardCard)
							zones.Player1.Discard++
						} else if entry.Player == 2 {
							zones.Player2.DiscardPile = append(zones.Player2.DiscardPile, discardCard)
							zones.Player2.Discard++
						}
					}
				}
			}
			
		case "quest":
			// Character quests, becomes exhausted (lore is handled by PlayerStatsLens)
			if entry.Parameters != nil {
				if cardID, ok := entry.Parameters["card_id"].(string); ok {
					// Find the battlefield card and exhaust it
					if entry.Player == 1 {
						for i := range zones.Player1.Battlefield {
							if zones.Player1.Battlefield[i].CardID == cardID {
								zones.Player1.Battlefield[i].Exhausted = true
								break
							}
						}
					} else if entry.Player == 2 {
						for i := range zones.Player2.Battlefield {
							if zones.Player2.Battlefield[i].CardID == cardID {
								zones.Player2.Battlefield[i].Exhausted = true
								break
							}
						}
					}
				}
			}
			
		case "pass":
			// Player passes their turn, next player draws a card (except P1's first turn)
			currentTurn++
			
			// Ready all battlefield cards for the next player
			nextPlayer := (currentTurn % 2) + 1
			if nextPlayer == 1 {
				for i := range zones.Player1.Battlefield {
					zones.Player1.Battlefield[i].Exhausted = false
				}
			} else {
				for i := range zones.Player2.Battlefield {
					zones.Player2.Battlefield[i].Exhausted = false
				}
			}
			
			// Draw a card unless it's player 1's first turn  
			// Note: In the new format, card drawing should be explicit draw_card actions
			// This logic is kept for compatibility but won't add actual cards to hand
			if !(nextPlayer == 1 && player1FirstTurn) {
				if nextPlayer == 1 && zones.Player1.Deck > 0 {
					zones.Player1.Deck--
					// In real game, a draw_card action would be logged here
				} else if nextPlayer == 2 && zones.Player2.Deck > 0 {
					zones.Player2.Deck--
					// In real game, a draw_card action would be logged here
				}
			}
			
			// After the first turn, P1 draws normally
			if nextPlayer == 1 {
				player1FirstTurn = false
			}
			
		case "turn_start":
			// New turn starting, update current turn from parameters
			if entry.Parameters != nil {
				switch v := entry.Parameters["turn"].(type) {
				case float64:
					currentTurn = int(v)
				case int:
					currentTurn = v
				}
			}
			
		case "challenge":
			// For challenges, we might need to handle damage/banishing
			// For now, just track the action without zone changes
			
		default:
			// Other actions don't affect zones for now
		}
	}
	
	// Convert to interface{} format for both available actions lens and frontend
	result := make(map[string]interface{})
	
	// Player 1 data (both formats for compatibility)
	player1Data := map[string]interface{}{
		"hand":        convertHandCards(zones.Player1.Hand),
		"deck":        zones.Player1.Deck,
		"discard":     zones.Player1.Discard,
		"discardPile": zones.Player1.DiscardPile,
		"ink":         zones.Player1.Ink,
		"in_play":     convertBattlefieldCards(zones.Player1.Battlefield),
		"battlefield": zones.Player1.Battlefield, // Also provide direct access
	}
	result["player_1"] = player1Data // For available actions lens
	result["player1"] = player1Data  // For frontend compatibility
	
	// Player 2 data (both formats for compatibility)
	player2Data := map[string]interface{}{
		"hand":        convertHandCards(zones.Player2.Hand),
		"deck":        zones.Player2.Deck,
		"discard":     zones.Player2.Discard,
		"discardPile": zones.Player2.DiscardPile,
		"ink":         zones.Player2.Ink,
		"in_play":     convertBattlefieldCards(zones.Player2.Battlefield),
		"battlefield": zones.Player2.Battlefield, // Also provide direct access
	}
	result["player_2"] = player2Data // For available actions lens
	result["player2"] = player2Data  // For frontend compatibility
	
	return result
}

// convertHandCards converts hand card IDs to interface{} format
func convertHandCards(handCards []string) []interface{} {
	result := make([]interface{}, len(handCards))
	for i, cardID := range handCards {
		result[i] = map[string]interface{}{
			"card_id": cardID,
		}
	}
	return result
}

// removeCardFromHand removes the first occurrence of cardID from hand
func removeCardFromHand(hand []string, cardID string) []string {
	for i, id := range hand {
		if id == cardID {
			// Remove card at index i
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand // Card not found, return original hand
}

// convertBattlefieldCards converts BattlefieldCard structs to interface{} format
func convertBattlefieldCards(battlefield []BattlefieldCard) []interface{} {
	result := make([]interface{}, len(battlefield))
	for i, card := range battlefield {
		result[i] = map[string]interface{}{
			"card_id":     card.CardID,
			"exhausted":   card.Exhausted,
			"id":          card.ID,
			"turn_played": card.TurnPlayed,
		}
	}
	return result
}