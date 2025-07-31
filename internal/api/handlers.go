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

// Global lens processor for API handlers
var lensProcessor *lens.Processor

func init() {
	lensProcessor = lens.New()
}

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
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
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
	
	actions, err := lensProcessor.Lens("availableActions", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute available actions: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, actions)
}

func GameTruncateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	var req struct {
		LogContent string `json:"logContent"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Load game from database to verify it exists
	_, err := game.LoadGameByID(gameID)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Update the game with the truncated log content
	if err := game.TruncateGameByID(gameID, req.LogContent); err != nil {
		writeError(w, fmt.Sprintf("failed to truncate game: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeResponse(w, map[string]string{"status": "success"})
}

func GameStepsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
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
	
	steps, err := lensProcessor.Lens("gameSteps", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute game steps: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, steps)
}

func GameStateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
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
	
	// Use composite lens to get all game state
	gameState, err := lensProcessor.Lens("composite", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute game state: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, gameState)
}

func GameBattlefieldHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
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
	
	battlefield, err := lensProcessor.Lens("battlefield", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute battlefield: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, battlefield)
}

func GameHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Parse log content from game
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to parse game log: %v", err), http.StatusInternalServerError)
		return
	}

	history, err := lensProcessor.Lens("history", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute game history: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, history)
}

func StepsNavigationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	
	// Load game from database
	gameData, err := game.LoadGameByID(gameID)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to load game: %v", err), http.StatusNotFound)
		return
	}
	
	// Parse log content from game
	entries, err := parser.ParseLogContent(gameData.LogContent)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to parse game log: %v", err), http.StatusInternalServerError)
		return
	}

	steps, err := lensProcessor.Lens("stepsNavigation", entries)
	if err != nil {
		writeError(w, fmt.Sprintf("failed to compute navigation steps: %v", err), http.StatusInternalServerError)
		return
	}
	writeResponse(w, steps)
}

func CacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := lensProcessor.CacheStats()
	writeResponse(w, stats)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{Error: message})
}