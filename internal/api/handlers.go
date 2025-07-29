package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"quards/internal/game"
	"quards/internal/lens"
	"quards/internal/parser"
)

type Response struct {
	Data interface{} `json:"data"`
	Error string     `json:"error,omitempty"`
}






func writeResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Data: data})
}


func GameAvailableActionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameName := vars["name"]
	
	// Load game from database
	gameData, err := game.LoadGameByName(gameName)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Parse log content from game
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to parse game log: %v", err), http.StatusBadRequest)
		return
	}
	
	// Check if step parameter is provided for historical context
	stepParam := r.URL.Query().Get("step")
	if stepParam != "" {
		if stepNum, err := strconv.Atoi(stepParam); err == nil && stepNum > 0 && stepNum <= len(entries) {
			// Only use entries up to the specified step (1-indexed)
			entries = entries[:stepNum]
		}
	}
	
	actions := lens.AvailableActionsLens(entries)
	writeResponse(w, actions)
}

func GameTruncateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameName := vars["name"]
	
	var req struct {
		LogContent string `json:"logContent"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Load game from database to verify it exists
	_, err := game.LoadGameByName(gameName)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Update the game with the truncated log content
	if err := game.TruncateGame(gameName, req.LogContent); err != nil {
		writeError(w, fmt.Sprintf("failed to truncate game: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, map[string]string{"status": "success"})
}

func GameByNameStepsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameName := vars["name"]
	
	// Load game from database
	gameData, err := game.LoadGameByName(gameName)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Parse log content from game
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to parse game log: %v", err), http.StatusBadRequest)
		return
	}
	
	steps := lens.GameStepsLens(entries)
	writeResponse(w, steps)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{Error: message})
}