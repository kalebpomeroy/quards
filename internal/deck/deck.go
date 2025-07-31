package deck

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	
	"quards/internal/database"
)

// Deck represents a constructed deck of 60 cards
type Deck struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Cards       map[string]int    `json:"cards"`       // CardID -> Count
	Created     time.Time         `json:"created"`
	Modified    time.Time         `json:"modified"`
}

// DeckList represents a simplified deck list for API responses
type DeckList struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CardCount   int       `json:"cardCount"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// ValidateDeck ensures the deck has exactly 60 cards
func (d *Deck) ValidateDeck() error {
	totalCards := 0
	for _, count := range d.Cards {
		totalCards += count
	}
	
	if totalCards != 60 {
		return fmt.Errorf("deck must have exactly 60 cards, got %d", totalCards)
	}
	
	return nil
}

// GetCardCount returns the total number of cards in the deck
func (d *Deck) GetCardCount() int {
	total := 0
	for _, count := range d.Cards {
		total += count
	}
	return total
}

// SaveDeck saves a deck to the database
func SaveDeck(deck *Deck) error {
	if err := deck.ValidateDeck(); err != nil {
		return err
	}
	
	db := database.GetDB()
	
	// Convert cards map to JSON
	cardsJSON, err := json.Marshal(deck.Cards)
	if err != nil {
		return fmt.Errorf("failed to marshal cards: %w", err)
	}
	
	// Check if deck exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM decks WHERE name = $1)", deck.Name).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check deck existence: %w", err)
	}
	
	if exists {
		// Update existing deck
		_, err = db.Exec(`
			UPDATE decks 
			SET description = $2, cards = $3, modified_at = NOW() 
			WHERE name = $1`,
			deck.Name, deck.Description, cardsJSON)
		if err != nil {
			return fmt.Errorf("failed to update deck: %w", err)
		}
	} else {
		// Insert new deck
		_, err = db.Exec(`
			INSERT INTO decks (name, description, cards) 
			VALUES ($1, $2, $3)`,
			deck.Name, deck.Description, cardsJSON)
		if err != nil {
			return fmt.Errorf("failed to insert deck: %w", err)
		}
	}
	
	return nil
}

// LoadDeck loads a deck from the database
func LoadDeck(name string) (*Deck, error) {
	db := database.GetDB()
	
	var deck Deck
	var cardsJSON []byte
	
	err := db.QueryRow(`
		SELECT id, name, description, cards, created_at, modified_at 
		FROM decks WHERE name = $1`, name).Scan(
		&deck.ID, &deck.Name, &deck.Description, &cardsJSON, &deck.Created, &deck.Modified)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deck not found: %s", name)
		}
		return nil, fmt.Errorf("failed to load deck: %w", err)
	}
	
	// Unmarshal cards JSON
	if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cards: %w", err)
	}
	
	return &deck, nil
}

// LoadDeckByID loads a deck from the database by ID
func LoadDeckByID(id int) (*Deck, error) {
	db := database.GetDB()
	
	var deck Deck
	var cardsJSON []byte
	
	err := db.QueryRow(`
		SELECT id, name, description, cards, created_at, modified_at 
		FROM decks WHERE id = $1`, id).Scan(
		&deck.ID, &deck.Name, &deck.Description, &cardsJSON, &deck.Created, &deck.Modified)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deck not found with ID: %d", id)
		}
		return nil, fmt.Errorf("failed to load deck: %w", err)
	}
	
	// Unmarshal cards JSON
	if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cards: %w", err)
	}
	
	return &deck, nil
}

// ListDecks returns a list of all available decks
func ListDecks() ([]DeckList, error) {
	db := database.GetDB()
	
	rows, err := db.Query(`
		SELECT id, name, description, cards, created_at, modified_at 
		FROM decks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query decks: %w", err)
	}
	defer rows.Close()
	
	var decks []DeckList
	for rows.Next() {
		var id int
		var name, description string
		var cardsJSON []byte
		var created, modified time.Time
		
		err := rows.Scan(&id, &name, &description, &cardsJSON, &created, &modified)
		if err != nil {
			continue // Skip invalid rows
		}
		
		// Parse cards to count them
		var cards map[string]int
		if err := json.Unmarshal(cardsJSON, &cards); err != nil {
			continue // Skip invalid cards
		}
		
		cardCount := 0
		for _, count := range cards {
			cardCount += count
		}
		
		decks = append(decks, DeckList{
			ID:          id,
			Name:        name,
			Description: description,
			CardCount:   cardCount,
			Created:     created,
			Modified:    modified,
		})
	}
	
	return decks, nil
}

// DeleteDeck removes a deck from the database
func DeleteDeck(name string) error {
	db := database.GetDB()
	
	result, err := db.Exec("DELETE FROM decks WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("deck not found: %s", name)
	}
	
	return nil
}

// DeleteDeckByID removes a deck from the database by ID
func DeleteDeckByID(id int) error {
	db := database.GetDB()
	
	result, err := db.Exec("DELETE FROM decks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("deck not found with ID: %d", id)
	}
	
	return nil
}