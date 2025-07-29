package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Turn       int                    `json:"turn"`
	Player     int                    `json:"player"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	GameState  map[string]interface{} `json:"game_state,omitempty"`
}


// ParseLogContent parses log entries from string content instead of file
func ParseLogContent(content string) ([]LogEntry, error) {
	if content == "" {
		return []LogEntry{}, nil
	}
	
	var entries []LogEntry
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse JSON on line %d: %w", lineNum+1, err)
		}
		
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// AppendLogEntry appends a new entry to a log file
func AppendLogEntry(filename string, entry LogEntry) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for append: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// CreateLogEntry creates a new log entry with current timestamp
func CreateLogEntry(turn, player int, action string, parameters map[string]interface{}) LogEntry {
	return LogEntry{
		Timestamp:  time.Now(),
		Turn:       turn,
		Player:     player,
		Action:     action,
		Parameters: parameters,
	}
}