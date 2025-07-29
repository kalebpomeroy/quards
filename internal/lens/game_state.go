package lens

import "quards/internal/parser"

// GameState represents overall game state information
type GameState struct {
	CurrentTurn int `json:"currentTurn"`
	TotalSteps  int `json:"totalSteps"`
}

// GameStateLens computes overall game state from log entries
func GameStateLens(entries []parser.LogEntry) interface{} {
	state := GameState{
		CurrentTurn: 0,
		TotalSteps:  len(entries),
	}
	
	for _, entry := range entries {
		if entry.Action == "pass" {
			state.CurrentTurn++
		} else if entry.Action == "turn_start" {
			// Update current turn from structured parameters
			if entry.Parameters != nil {
				switch v := entry.Parameters["turn"].(type) {
				case float64:
					state.CurrentTurn = int(v)
				case int:
					state.CurrentTurn = v
				}
			}
		}
	}
	
	return state
}