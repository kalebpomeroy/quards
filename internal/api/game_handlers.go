package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"quards/internal/game"
)

// CreateGameHandler creates a new game
func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	var req game.CreateGameRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Player1Deck == "" || req.Player2Deck == "" {
		writeError(w, "both player1Deck and player2Deck are required", http.StatusBadRequest)
		return
	}
	
	createdGame, err := game.CreateGame(&req)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to create game: %v", err), http.StatusBadRequest)
		return
	}
	
	writeResponse(w, createdGame)
}

// ListGamesHandler returns all games, optionally filtered by deck
func ListGamesHandler(w http.ResponseWriter, r *http.Request) {
	deckFilter := r.URL.Query().Get("deck")
	
	games, err := game.ListGames(deckFilter)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to list games: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, games)
}

// GetGameHandler returns a specific game by ID
func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	
	gameData, err := game.LoadGame(gameID)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	writeResponse(w, gameData)
}


// DeleteGameHandler deletes a game
func DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	
	if err := game.DeleteGame(gameID); err != nil {
		writeError(w, fmt.Sprintf("failed to delete game: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, map[string]string{"message": "game deleted successfully"})
}

// ExecuteActionRequest represents a request to execute an action in a game
type ExecuteActionRequest struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ExecuteActionHandler executes an action in a game and appends it to the log
func ExecuteActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	var req ExecuteActionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Type == "" {
		writeError(w, "action type is required", http.StatusBadRequest)
		return
	}
	
	// Execute the action by appending to the game log
	err := game.AppendActionToGameByID(gameID, req.Type, req.Parameters)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to execute action: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, map[string]string{"message": "action executed successfully"})
}