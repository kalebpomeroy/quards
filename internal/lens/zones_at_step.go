package lens

import (
	"quards/internal/parser"
)

// ZonesAtStepLens returns zones state after executing up to a specific step
func ZonesAtStepLens(entries []parser.LogEntry, step int) interface{} {
	// Limit entries to the step we want
	if step >= len(entries) {
		step = len(entries) - 1
	}
	if step < 0 {
		step = 0
	}
	
	limitedEntries := entries[:step+1]
	return ZonesLens(limitedEntries)
}

// GameStepsLens returns information about each step in the game
func GameStepsLens(entries []parser.LogEntry) interface{} {
	steps := make([]map[string]interface{}, len(entries))
	
	for i, entry := range entries {
		steps[i] = map[string]interface{}{
			"step":       i,
			"player":     entry.Player,
			"action":     entry.Action,
			"parameters": entry.Parameters,
			"gameState":  CompositeGameStateAtStepLens(entries, i),
		}
	}
	
	return steps
}