package lens

import (
	"quards/internal/parser"
	"regexp"
	"fmt"
)

// CardData represents the structure of card data from the JSON file
type CardData struct {
	Unique_ID    string `json:"Unique_ID"`
	Name         string `json:"Name"`
	Type         string `json:"Type"`
	Color        string `json:"Color"`
	Cost         int    `json:"Cost"`
	Inkable      bool   `json:"Inkable"`
	Strength     int    `json:"Strength"`
	Willpower    int    `json:"Willpower"`
	Lore         int    `json:"Lore"`
	Rarity       string `json:"Rarity"`
	Body_Text    string `json:"Body_Text"`
	Flavor_Text  string `json:"Flavor_Text"`
	Image        string `json:"Image"`
}

// LoadCardDatabase is deprecated - use GetCardDatabase() instead
// Kept for backward compatibility during transition
func LoadCardDatabase() map[string]CardData {
	return GetCardDatabase()
}

// CardNamesLens replaces card IDs with card names in the log entries
func CardNamesLens(entries []parser.LogEntry) interface{} {
	cardDB := LoadCardDatabase()
	
	result := make([]map[string]interface{}, len(entries))
	
	for i, entry := range entries {
		// Convert parameters to string for translation
		paramStr := ""
		if entry.Parameters != nil {
			// Try to extract card IDs from structured parameters
			if cardID, ok := entry.Parameters["card_id"].(string); ok {
				paramStr = cardID
			}
		}
		
		result[i] = map[string]interface{}{
			"player":     entry.Player,
			"action":     entry.Action,
			"parameters": translateCardIDs(paramStr, cardDB),
			"original":   entry.Parameters,
		}
	}
	
	return result
}

// translateCardIDs finds card IDs in parameters and replaces them with names
func translateCardIDs(params string, cardDB map[string]CardData) string {
	if params == "" {
		return params
	}
	
	// Match card IDs like "ARI-001", "28", etc.
	// First try set-number format (ARI-001)
	setCardRegex := regexp.MustCompile(`\b[A-Z]{2,4}-\d{3}\b`)
	result := setCardRegex.ReplaceAllStringFunc(params, func(cardID string) string {
		if card, exists := cardDB[cardID]; exists {
			return fmt.Sprintf("%s (%s)", card.Name, cardID)
		}
		return cardID
	})
	
	return result
}