package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	
	"github.com/gorilla/mux"
	"quards/internal/deck"
)

// ListDecksHandler returns all available decks
func ListDecksHandler(w http.ResponseWriter, r *http.Request) {
	decks, err := deck.ListDecks()
	if err != nil {
		writeError(w, fmt.Sprintf("failed to list decks: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, decks)
}

// GetDeckHandler returns a specific deck by ID
func GetDeckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "invalid deck ID", http.StatusBadRequest)
		return
	}
	
	deckData, err := deck.LoadDeckByID(id)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load deck: %v", err), http.StatusNotFound)
		return
	}
	
	writeResponse(w, deckData)
}

// GetDeckByNameHandler returns a specific deck by name (legacy support)
func GetDeckByNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deckName := vars["deckname"]
	
	// Security: basic path validation
	if filepath.Base(deckName) != deckName {
		writeError(w, "invalid deck name", http.StatusBadRequest)
		return
	}
	
	deckData, err := deck.LoadDeck(deckName)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load deck: %v", err), http.StatusNotFound)
		return
	}
	
	writeResponse(w, deckData)
}

// CreateDeckHandler creates a new deck
func CreateDeckHandler(w http.ResponseWriter, r *http.Request) {
	var deckData deck.Deck
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&deckData); err != nil {
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := deck.SaveDeck(&deckData); err != nil {
		writeError(w, fmt.Sprintf("failed to save deck: %v", err), http.StatusBadRequest)
		return
	}
	
	writeResponse(w, map[string]string{"message": "deck created successfully"})
}

// UpdateDeckHandler updates an existing deck
func UpdateDeckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deckName := vars["deckname"]
	
	// Security: basic path validation
	if filepath.Base(deckName) != deckName {
		writeError(w, "invalid deck name", http.StatusBadRequest)
		return
	}
	
	var deckData deck.Deck
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&deckData); err != nil {
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Ensure the deck name matches the URL
	deckData.Name = deckName
	
	if err := deck.SaveDeck(&deckData); err != nil {
		writeError(w, fmt.Sprintf("failed to save deck: %v", err), http.StatusBadRequest)
		return
	}
	
	writeResponse(w, map[string]string{"message": "deck updated successfully"})
}

// DeleteDeckHandler deletes a deck
func DeleteDeckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "invalid deck ID", http.StatusBadRequest)
		return
	}
	
	if err := deck.DeleteDeckByID(id); err != nil {
		writeError(w, fmt.Sprintf("failed to delete deck: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, map[string]string{"message": "deck deleted successfully"})
}