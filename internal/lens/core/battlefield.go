package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// BattlefieldCharacter represents a character on the battlefield with full state
type BattlefieldCharacter struct {
	CardID      string            `json:"card_id"`
	Owner       int               `json:"owner"`       // Which player owns this character
	Exhausted   bool              `json:"exhausted"`   // Whether the character is exhausted
	TurnPlayed  int               `json:"turn_played"` // Which turn this character was played
	Damage      int               `json:"damage"`      // Current damage on the character
	Counters    map[string]int    `json:"counters"`    // Various counters (lore, strength, etc.)
	Abilities   []string          `json:"abilities"`   // Active abilities
	Attachments []BattlefieldItem `json:"attachments"` // Items attached to this character
	Position    int               `json:"position"`    // Position on battlefield (for ordering)

	// Computed fields from card database
	Name          string `json:"name"`
	BaseWillpower int    `json:"base_willpower"`
	BaseStrength  int    `json:"base_strength"`
	BaseLore      int    `json:"base_lore"`
	CardType      string `json:"card_type"`
}

// BattlefieldItem represents an item on the battlefield
type BattlefieldItem struct {
	CardID     string         `json:"card_id"`
	Owner      int            `json:"owner"`
	TurnPlayed int            `json:"turn_played"`
	AttachedTo string         `json:"attached_to,omitempty"` // CardID if attached to character
	Counters   map[string]int `json:"counters"`
	Position   int            `json:"position"`

	// Computed fields
	Name     string `json:"name"`
	CardType string `json:"card_type"`
}

// BattlefieldState represents the complete battlefield state
type BattlefieldState struct {
	Characters   []BattlefieldCharacter `json:"characters"`
	Items        []BattlefieldItem      `json:"items"`
	Turn         int                    `json:"turn"`
	ActivePlayer int                    `json:"active_player"`
}

// BattlefieldLens computes the complete battlefield state (pure function)
func BattlefieldLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	battlefield := BattlefieldState{
		Characters:   []BattlefieldCharacter{},
		Items:        []BattlefieldItem{},
		Turn:         1,
		ActivePlayer: 1,
	}

	cardDB := services.CardDB.GetAll()
	nextPosition := 1

	for _, entry := range entries {
		switch entry.Event {
		case parser.TurnStarted:
			battlefield.Turn = entry.GetInt("turn")
			battlefield.ActivePlayer = entry.GetPlayer()

		case parser.CardPlayed:
			cardID := entry.GetCard("card")
			_ = entry.GetInstance("instance") // TODO: Use instance ID for tracking
			player := entry.GetPlayer()

			// Check if this is a character or item
			if cardData, exists := cardDB[cardID]; exists {
				switch cardData.Type {
				case "Character":
					character := BattlefieldCharacter{
						CardID:        cardID,
						Owner:         player,
						Exhausted:     false, // Characters are not exhausted when played
						TurnPlayed:    battlefield.Turn,
						Damage:        0,
						Counters:      make(map[string]int),
						Abilities:     []string{},
						Attachments:   []BattlefieldItem{},
						Position:      nextPosition,
						Name:          cardData.Name,
						BaseWillpower: cardData.Willpower,
						BaseStrength:  cardData.Strength,
						BaseLore:      cardData.Lore,
						CardType:      cardData.Type,
					}
					battlefield.Characters = append(battlefield.Characters, character)
					nextPosition++

				case "Item":
					item := BattlefieldItem{
						CardID:     cardID,
						Owner:      player,
						TurnPlayed: battlefield.Turn,
						Counters:   make(map[string]int),
						Position:   nextPosition,
						Name:       cardData.Name,
						CardType:   cardData.Type,
					}

					// Check if item is being attached to a character
					targetID := entry.GetCard("target")
					if targetID != "" {
						item.AttachedTo = targetID
						// Add to character's attachments
						for i := range battlefield.Characters {
							if battlefield.Characters[i].CardID == targetID {
								battlefield.Characters[i].Attachments = append(
									battlefield.Characters[i].Attachments, item)
								break
							}
						}
					} else {
						battlefield.Items = append(battlefield.Items, item)
					}
					nextPosition++
				}
			}

		case parser.QuestAttempted:
			_ = entry.GetInstance("instance") // TODO: Use instance ID for matching
			player := entry.GetPlayer()

			// Mark character as exhausted after questing by instance ID
			for i := range battlefield.Characters {
				if battlefield.Characters[i].Owner == player {
					// For now, use CardID matching until we add InstanceID to BattlefieldCharacter
					// TODO: Add InstanceID field to BattlefieldCharacter
					battlefield.Characters[i].Exhausted = true
					break
				}
			}

		case parser.TurnPassed:
			player := entry.GetPlayer()

			// On pass, reset exhausted status for all characters of the passing player
			for i := range battlefield.Characters {
				if battlefield.Characters[i].Owner == player {
					battlefield.Characters[i].Exhausted = false
				}
			}
		}
	}

	return battlefield
}
