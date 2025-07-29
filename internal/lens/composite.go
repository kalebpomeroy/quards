package lens

import "quards/internal/parser"

// CompositeGameState combines all lens outputs into a single response
type CompositeGameState struct {
	Zones       interface{} `json:"zones"`
	PlayerStats interface{} `json:"playerStats"`
	GameState   interface{} `json:"gameState"`
}

// CompositeGameStateLens combines multiple lenses into a single output
func CompositeGameStateLens(entries []parser.LogEntry) interface{} {
	return CompositeGameState{
		Zones:       ZonesLens(entries),
		PlayerStats: PlayerStatsLens(entries),
		GameState:   GameStateLens(entries),
	}
}

// CompositeGameStateAtStepLens returns composite state at a specific step
func CompositeGameStateAtStepLens(entries []parser.LogEntry, step int) interface{} {
	// Limit entries to the step we want
	if step >= len(entries) {
		step = len(entries) - 1
	}
	if step < 0 {
		step = 0
	}
	
	limitedEntries := entries[:step+1]
	return CompositeGameStateLens(limitedEntries)
}