package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// CompositeGameState represents the combined output of multiple lenses
type CompositeGameState struct {
	Zones       interface{} `json:"zones"`
	PlayerStats interface{} `json:"playerStats"`
	Battlefield interface{} `json:"battlefield"`
	GameState   interface{} `json:"gameState,omitempty"`
}

// CompositeGameStateLens combines multiple lenses into a single output (pure function)
func CompositeGameStateLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	return CompositeGameState{
		Zones:       ZonesLens(entries, services),
		PlayerStats: PlayerStatsLens(entries, services),
		Battlefield: BattlefieldLens(entries, services),
		GameState:   nil, // Placeholder for future composite state
	}
}