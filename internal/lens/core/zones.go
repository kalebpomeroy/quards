package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// HandCard represents a card in a player's hand
type HandCard struct {
	CardID string `json:"card_id"`
}

// InPlayCard represents a card in play with instance tracking
type InPlayCard struct {
	CardID     string `json:"card_id"`
	InstanceID string `json:"instance_id"` // Battlefield instance like $CHAR_001
	Exhausted  bool   `json:"exhausted"`
	TurnPlayed int    `json:"turn_played"`
}

// InkCard represents a card in the inkwell
type InkCard struct {
	CardID string `json:"card_id"`
}

// ZonesLens computes zones state from event-sourcing log format
func ZonesLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	zones := map[string]interface{}{
		"player1": map[string]interface{}{
			"hand":    []HandCard{},
			"in_play": []InPlayCard{},
			"ink":     []InkCard{},
			"deck":    0,
			"discard": 0,
		},
		"player2": map[string]interface{}{
			"hand":    []HandCard{},
			"in_play": []InPlayCard{},
			"ink":     []InkCard{},
			"deck":    0,
			"discard": 0,
		},
	}

	currentTurn := 1

	for _, entry := range entries {
		switch entry.Event {
		case parser.GameStarted:
			// Initialize deck counts (60 cards each, standard deck size)
			player1Zones := zones["player1"].(map[string]interface{})
			player2Zones := zones["player2"].(map[string]interface{})
			
			player1Zones["deck"] = 60
			player2Zones["deck"] = 60

		case parser.TurnStarted:
			currentTurn = entry.GetInt("turn")

		case parser.OpeningHandsDrawn:
			// Player 1 opening hand
			p1Cards := entry.GetStringSlice("p1")
			player1Zones := zones["player1"].(map[string]interface{})
			hand := make([]HandCard, len(p1Cards))
			for i, cardID := range p1Cards {
				hand[i] = HandCard{CardID: cardID}
			}
			player1Zones["hand"] = hand
			
			// Remove drawn cards from deck count
			p1DeckCount := player1Zones["deck"].(int)
			player1Zones["deck"] = p1DeckCount - len(p1Cards)

			// Player 2 opening hand
			p2Cards := entry.GetStringSlice("p2")
			player2Zones := zones["player2"].(map[string]interface{})
			hand2 := make([]HandCard, len(p2Cards))
			for i, cardID := range p2Cards {
				hand2[i] = HandCard{CardID: cardID}
			}
			player2Zones["hand"] = hand2
			
			// Remove drawn cards from deck count
			p2DeckCount := player2Zones["deck"].(int)
			player2Zones["deck"] = p2DeckCount - len(p2Cards)

		case parser.CardDrawn:
			cardID := entry.GetCard("card")
			player := entry.GetPlayer()
			playerKey := getPlayerZoneKey(player)
			
			if playerZones, ok := zones[playerKey].(map[string]interface{}); ok {
				hand := playerZones["hand"].([]HandCard)
				hand = append(hand, HandCard{CardID: cardID})
				playerZones["hand"] = hand
				
				// Remove one card from deck count
				deckCount := playerZones["deck"].(int)
				if deckCount > 0 {
					playerZones["deck"] = deckCount - 1
				}
			}

		case parser.CardPlayed:
			cardID := entry.GetCard("card_id")
			instanceID := entry.GetInstance("instance")
			player := entry.GetPlayer()
			playerKey := getPlayerZoneKey(player)
			
			if playerZones, ok := zones[playerKey].(map[string]interface{}); ok {
				// Remove from hand
				hand := playerZones["hand"].([]HandCard)
				newHand := make([]HandCard, 0, len(hand))
				for _, card := range hand {
					if card.CardID != cardID {
						newHand = append(newHand, card)
					}
				}
				playerZones["hand"] = newHand

				// Check card type to determine destination
				if cardData, exists := services.CardDB.GetCard(cardID); exists {
					if cardData.Type == "Action" {
						// Actions go to discard pile
						discardCount := playerZones["discard"].(int)
						playerZones["discard"] = discardCount + 1
					} else {
						// Characters, Items, Locations go to in_play
						inPlay := playerZones["in_play"].([]InPlayCard)
						inPlay = append(inPlay, InPlayCard{
							CardID:     cardID,
							InstanceID: string(instanceID),
							Exhausted:  false,
							TurnPlayed: currentTurn,
						})
						playerZones["in_play"] = inPlay
					}
				} else {
					// Fallback: if card not found in database, assume it goes to in_play
					inPlay := playerZones["in_play"].([]InPlayCard)
					inPlay = append(inPlay, InPlayCard{
						CardID:     cardID,
						InstanceID: string(instanceID),
						Exhausted:  false,
						TurnPlayed: currentTurn,
					})
					playerZones["in_play"] = inPlay
				}
			}

		case parser.CardInked:
			cardID := entry.GetCard("card_id")
			player := entry.GetPlayer()
			playerKey := getPlayerZoneKey(player)
			
			if playerZones, ok := zones[playerKey].(map[string]interface{}); ok {
				// Remove from hand
				hand := playerZones["hand"].([]HandCard)
				newHand := make([]HandCard, 0, len(hand))
				for _, card := range hand {
					if card.CardID != cardID {
						newHand = append(newHand, card)
					}
				}
				playerZones["hand"] = newHand

				// Add to ink
				ink := playerZones["ink"].([]InkCard)
				ink = append(ink, InkCard{CardID: cardID})
				playerZones["ink"] = ink
			}

		case parser.QuestAttempted:
			instanceID := entry.GetInstance("instance")
			player := entry.GetPlayer()
			playerKey := getPlayerZoneKey(player)
			
			if playerZones, ok := zones[playerKey].(map[string]interface{}); ok {
				// Mark character as exhausted by instance ID
				inPlay := playerZones["in_play"].([]InPlayCard)
				for i := range inPlay {
					if inPlay[i].InstanceID == string(instanceID) {
						inPlay[i].Exhausted = true
						break
					}
				}
				playerZones["in_play"] = inPlay
			}

		case parser.CharacterExerted:
			instanceID := entry.GetInstance("instance")
			// Find which player owns this instance
			for _, playerZones := range zones {
				if pz, ok := playerZones.(map[string]interface{}); ok {
					inPlay := pz["in_play"].([]InPlayCard)
					for i := range inPlay {
						if inPlay[i].InstanceID == string(instanceID) {
							inPlay[i].Exhausted = true
							pz["in_play"] = inPlay
							goto nextEntry
						}
					}
				}
			}

		case parser.CharacterReadied:
			instanceID := entry.GetInstance("instance")
			// Find which player owns this instance
			for _, playerZones := range zones {
				if pz, ok := playerZones.(map[string]interface{}); ok {
					inPlay := pz["in_play"].([]InPlayCard)
					for i := range inPlay {
						if inPlay[i].InstanceID == string(instanceID) {
							inPlay[i].Exhausted = false
							pz["in_play"] = inPlay
							goto nextEntry
						}
					}
				}
			}

		case parser.CharacterBanished:
			instanceID := entry.GetInstance("instance")
			// Remove character from battlefield by instance ID
			for _, playerZones := range zones {
				if pz, ok := playerZones.(map[string]interface{}); ok {
					inPlay := pz["in_play"].([]InPlayCard)
					for i, card := range inPlay {
						if card.InstanceID == string(instanceID) {
							// Remove character from in_play
							newInPlay := append(inPlay[:i], inPlay[i+1:]...)
							pz["in_play"] = newInPlay
							
							// Add to discard pile
							discardCount := pz["discard"].(int)
							pz["discard"] = discardCount + 1
							goto nextEntry
						}
					}
				}
			}

		case parser.TurnPassed:
			player := entry.GetPlayer()
			playerKey := getPlayerZoneKey(player)
			
			if playerZones, ok := zones[playerKey].(map[string]interface{}); ok {
				// Ready all characters for the player whose turn is ending
				inPlay := playerZones["in_play"].([]InPlayCard)
				for i := range inPlay {
					inPlay[i].Exhausted = false
				}
				playerZones["in_play"] = inPlay
			}
		}
		nextEntry:
	}

	// Convert typed structs to generic interface{} format for JSON compatibility
	result := make(map[string]interface{})
	for playerKey, playerZones := range zones {
		pz := playerZones.(map[string]interface{})
		convertedZones := make(map[string]interface{})
		
		for zoneKey, zoneData := range pz {
			switch zoneKey {
			case "hand":
				handCards := zoneData.([]HandCard)
				handInterface := make([]interface{}, len(handCards))
				for i, card := range handCards {
					handInterface[i] = map[string]interface{}{
						"card_id": card.CardID,
					}
				}
				convertedZones[zoneKey] = handInterface
			case "in_play":
				inPlayCards := zoneData.([]InPlayCard)
				inPlayInterface := make([]interface{}, len(inPlayCards))
				for i, card := range inPlayCards {
					inPlayInterface[i] = map[string]interface{}{
						"card_id":     card.CardID,
						"instance_id": card.InstanceID,
						"exhausted":   card.Exhausted,
						"turn_played": card.TurnPlayed,
					}
				}
				convertedZones[zoneKey] = inPlayInterface
			case "ink":
				inkCards := zoneData.([]InkCard)
				inkInterface := make([]interface{}, len(inkCards))
				for i, card := range inkCards {
					inkInterface[i] = map[string]interface{}{
						"card_id": card.CardID,
					}
				}
				convertedZones[zoneKey] = inkInterface
			default:
				convertedZones[zoneKey] = zoneData
			}
		}
		result[playerKey] = convertedZones
	}
	
	return result
}

// Helper function to get the correct player zone key
func getPlayerZoneKey(player int) string {
	if player == 1 {
		return "player1"
	}
	return "player2"
}