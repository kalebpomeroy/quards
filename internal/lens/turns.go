package lens

import "quards/internal/parser"

// TurnsLens counts the number of turns in a game
// A turn is any sequence of player actions ending with PASS (including just PASS)
func TurnsLens(entries []parser.LogEntry) interface{} {
	if len(entries) == 0 {
		return 0
	}
	
	turns := 0
	
	for _, entry := range entries {
		// Skip system actions
		if entry.Player == 0 {
			continue
		}
		
		// Every pass marks the end of a turn
		if entry.Action == "pass" {
			turns++
		}
	}
	
	return turns
}