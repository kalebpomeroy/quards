package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// GameStateLens provides current game state information (pure function)
func GameStateLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	return map[string]interface{}{
		"currentPlayer": getCurrentPlayer(entries),
		"currentTurn":   getCurrentTurn(entries),
	}
}