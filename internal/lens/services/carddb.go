package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// InMemoryCardDB implements CardDatabase interface with in-memory storage
type InMemoryCardDB struct {
	cards map[string]*CardData
	mutex sync.RWMutex
	loaded bool
}

// NewInMemoryCardDB creates a new in-memory card database
func NewInMemoryCardDB() *InMemoryCardDB {
	return &InMemoryCardDB{
		cards: make(map[string]*CardData),
	}
}

// LoadFromFile loads card data from a JSON file
func (db *InMemoryCardDB) LoadFromFile(filePath string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open card database file: %w", err)
	}
	defer file.Close()

	var cards []CardData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cards); err != nil {
		return fmt.Errorf("failed to decode card database: %w", err)
	}

	// Clear existing cards and load new ones
	db.cards = make(map[string]*CardData)
	for i := range cards {
		card := &cards[i]
		db.cards[card.UniqueID] = card
	}

	db.loaded = true
	return nil
}

// GetCard retrieves a card by its unique ID
func (db *InMemoryCardDB) GetCard(id string) (*CardData, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	card, exists := db.cards[id]
	return card, exists
}

// GetAll returns all cards in the database
func (db *InMemoryCardDB) GetAll() map[string]*CardData {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make(map[string]*CardData, len(db.cards))
	for k, v := range db.cards {
		result[k] = v
	}
	return result
}

// IsLoaded returns whether the database has been loaded
func (db *InMemoryCardDB) IsLoaded() bool {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.loaded
}

// Count returns the number of cards in the database
func (db *InMemoryCardDB) Count() int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return len(db.cards)
}