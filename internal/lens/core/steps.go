package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// GameStepsLens returns information about each step in the game (pure function)
func GameStepsLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	steps := make([]map[string]interface{}, len(entries))

	for i, entry := range entries {
		steps[i] = map[string]interface{}{
			"step":       entry.Step,
			"player":     entry.GetPlayer(),
			"event":      string(entry.Event),
			"parameters": entry.Parameters,
			// Remove expensive gameState calculation - frontend can request specific steps if needed
		}
	}

	return steps
}