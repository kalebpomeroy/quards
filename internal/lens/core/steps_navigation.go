package core

import (
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// NavigationStep represents a step optimized for UI navigation
type NavigationStep struct {
	Step            int               `json:"step"`
	Player          int               `json:"player"`
	Event           string            `json:"event"`
	Action          string            `json:"action"`          // UI-friendly action name
	Description     string            `json:"description"`     // Human-readable description
	Parameters      map[string]string `json:"parameters"`
	IsPlayerChoice  bool              `json:"isPlayerChoice"`  // True if this was a player decision
	IsFramework     bool              `json:"isFramework"`     // True if this was automatic/system
}

// StepsNavigationLens provides step data optimized for UI navigation and history display
func StepsNavigationLens(entries []parser.LogEntry, services *services.LensServices) interface{} {
	steps := make([]NavigationStep, len(entries))

	for i, entry := range entries {
		step := NavigationStep{
			Step:           entry.Step,
			Player:         entry.GetPlayer(),
			Event:          string(entry.Event),
			Action:         mapEventToAction(entry.Event),
			Description:    formatEventDescription(entry, services.CardDB),
			Parameters:     entry.Parameters,
			IsPlayerChoice: isPlayerChoiceEvent(entry.Event),
			IsFramework:    !isPlayerChoiceEvent(entry.Event),
		}

		steps[i] = step
	}

	return steps
}

// mapEventToAction converts event names to UI-friendly action names
func mapEventToAction(event parser.LogEventType) string {
	switch event {
	case parser.GameStarted:
		return "game_start"
	case parser.DecksShuffled:
		return "shuffle_decks"
	case parser.OpeningHandsDrawn:
		return "draw_opening_hands"
	case parser.TurnStarted:
		return "turn_start"
	case parser.CardDrawn:
		return "draw_card"
	case parser.CardInked:
		return "ink_card"
	case parser.CardPlayed:
		return "play_card"
	case parser.QuestAttempted:
		return "quest"
	case parser.TurnPassed:
		return "pass"
	case parser.CharacterExerted:
		return "exert_character"
	case parser.CharacterReadied:
		return "ready_character"
	case parser.CharacterBanished:
		return "banish_character"
	default:
		return string(event)
	}
}

// isPlayerChoiceEvent determines if an event represents a player choice vs framework action
func isPlayerChoiceEvent(event parser.LogEventType) bool {
	playerChoiceEvents := []parser.LogEventType{
		parser.CardInked,
		parser.CardPlayed,
		parser.QuestAttempted,
		parser.TurnPassed,
		// Add other player choice events as needed
	}

	for _, choice := range playerChoiceEvents {
		if event == choice {
			return true
		}
	}
	return false
}