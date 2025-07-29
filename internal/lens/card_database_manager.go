package lens

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Global card database singleton
var (
	globalCardDB     map[string]CardData
	cardDBOnce       sync.Once
	cardDBInitialized bool
)

// InitializeCardDatabase loads the card database once at application startup
func InitializeCardDatabase() error {
	var initErr error
	cardDBOnce.Do(func() {
		log.Println("Loading card database...")
		
		file, err := os.Open("static/cards.json")
		if err != nil {
			initErr = err
			return
		}
		defer file.Close()
		
		var cards []CardData
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&cards); err != nil {
			initErr = err
			return
		}
		
		// Create lookup map by Unique_ID
		globalCardDB = make(map[string]CardData)
		for _, card := range cards {
			globalCardDB[card.Unique_ID] = card
		}
		
		cardDBInitialized = true
		log.Printf("Loaded %d cards into database", len(cards))
	})
	
	return initErr
}

// GetCardDatabase returns the global card database
// Panics if not initialized - ensures proper startup sequence
func GetCardDatabase() map[string]CardData {
	if !cardDBInitialized {
		panic("Card database not initialized. Call InitializeCardDatabase() at application startup.")
	}
	return globalCardDB
}

// IsCardDatabaseInitialized returns whether the card database has been loaded
func IsCardDatabaseInitialized() bool {
	return cardDBInitialized
}