package parser

import (
	"strconv"
	"strings"
)

// LogEventType represents the type of game event
type LogEventType string

// Event types for the log format
const (
	GameStarted             LogEventType = "GameStarted"
	DecksShuffled           LogEventType = "DecksShuffled"
	OpeningHandsDrawn       LogEventType = "OpeningHandsDrawn"
	TurnStarted             LogEventType = "TurnStarted"
	CardDrawn               LogEventType = "CardDrawn"
	CardInked               LogEventType = "CardInked"
	CardPlayed              LogEventType = "CardPlayed"
	ItemPlayed              LogEventType = "ItemPlayed"
	LocationPlayed          LogEventType = "LocationPlayed"
	QuestAttempted          LogEventType = "QuestAttempted"
	CharacterExerted        LogEventType = "CharacterExerted"
	CharacterReadied        LogEventType = "CharacterReadied"
	CharacterBanished       LogEventType = "CharacterBanished"
	ItemAttached            LogEventType = "ItemAttached"
	ItemDetached            LogEventType = "ItemDetached"
	ItemBanished            LogEventType = "ItemBanished"
	LocationEffectTriggered LogEventType = "LocationEffectTriggered"
	LocationDestroyed       LogEventType = "LocationDestroyed"
	CounterAdded            LogEventType = "CounterAdded"
	CounterRemoved          LogEventType = "CounterRemoved"
	TurnPassed              LogEventType = "TurnPassed"
)

// InstanceID represents a battlefield object instance
type InstanceID string

// Type returns the type of instance (character, item, location)
func (id InstanceID) Type() string {
	s := string(id)
	if strings.HasPrefix(s, "$CHAR_") {
		return "character"
	} else if strings.HasPrefix(s, "$ITEM_") {
		return "item"
	} else if strings.HasPrefix(s, "$LOC_") {
		return "location"
	}
	return "unknown"
}

// IsValid checks if the instance ID has valid format
func (id InstanceID) IsValid() bool {
	s := string(id)
	return strings.HasPrefix(s, "$CHAR_") ||
		strings.HasPrefix(s, "$ITEM_") ||
		strings.HasPrefix(s, "$LOC_")
}

// LogEntry represents a single event in the game log
type LogEntry struct {
	Event      LogEventType
	Parameters map[string]string
	Step       int
}

// GetPlayer returns the player number from the event parameters
func (e *LogEntry) GetPlayer() int {
	if playerStr, ok := e.Parameters["player"]; ok {
		if player, err := strconv.Atoi(playerStr); err == nil {
			return player
		}
	}
	return 0 // System event
}

// GetInstance returns an instance ID from the parameters
func (e *LogEntry) GetInstance(key string) InstanceID {
	if instanceStr, ok := e.Parameters[key]; ok {
		return InstanceID(instanceStr)
	}
	return ""
}

// GetCard returns a card ID from the parameters
func (e *LogEntry) GetCard(key string) string {
	if cardStr, ok := e.Parameters[key]; ok {
		return cardStr
	}
	return ""
}

// GetInt returns an integer value from the parameters
func (e *LogEntry) GetInt(key string) int {
	if valueStr, ok := e.Parameters[key]; ok {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return 0
}

// GetStringSlice returns a comma-separated string as a slice
func (e *LogEntry) GetStringSlice(key string) []string {
	if valueStr, ok := e.Parameters[key]; ok {
		// Remove quotes and split by comma
		cleaned := strings.Trim(valueStr, `"`)
		if cleaned == "" {
			return []string{}
		}
		return strings.Split(cleaned, ",")
	}
	return []string{}
}

// ParseLogContent parses the event-sourcing log format
func ParseLogContent(content string) ([]LogEntry, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	events := make([]LogEntry, 0, len(lines))
	step := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		event, err := parseLogLine(line, step)
		if err != nil {
			return nil, err
		}

		events = append(events, *event)
		step++
	}

	return events, nil
}

// parseLogLine parses a single event line
func parseLogLine(line string, step int) (*LogEntry, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, nil
	}

	entry := &LogEntry{
		Event:      LogEventType(parts[0]),
		Parameters: make(map[string]string),
		Step:       step,
	}

	// Parse key=value pairs
	for _, part := range parts[1:] {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue // Skip malformed parameters
		}

		// Remove surrounding quotes if present
		value = strings.Trim(value, `"`)
		entry.Parameters[key] = value

	}

	return entry, nil
}
