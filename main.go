package main

import (
	"fmt"
	"log"
	"net/http"
	
	"github.com/gorilla/mux"
	"quards/internal/api"
	"quards/internal/database"
	"quards/internal/lens"
)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.CloseDB()
	
	// Initialize card database at startup
	if err := lens.InitializeCardDatabase(); err != nil {
		log.Fatal("Failed to initialize card database:", err)
	}
	
	r := mux.NewRouter()
	
	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	
	// Deck management endpoints
	apiRouter.HandleFunc("/decks", api.ListDecksHandler).Methods("GET")
	apiRouter.HandleFunc("/decks", api.CreateDeckHandler).Methods("POST")
	apiRouter.HandleFunc("/decks/{deckname}", api.GetDeckHandler).Methods("GET")
	apiRouter.HandleFunc("/decks/{deckname}", api.UpdateDeckHandler).Methods("PUT")
	apiRouter.HandleFunc("/decks/{deckname}", api.DeleteDeckHandler).Methods("DELETE")
	
	// Game management endpoints
	apiRouter.HandleFunc("/games", api.ListGamesHandler).Methods("GET")
	apiRouter.HandleFunc("/games", api.CreateGameHandler).Methods("POST")
	apiRouter.HandleFunc("/games/{id}", api.GetGameHandler).Methods("GET")
	apiRouter.HandleFunc("/games/{id}", api.DeleteGameHandler).Methods("DELETE")
	apiRouter.HandleFunc("/games/by-name/{name}", api.GetGameByNameHandler).Methods("GET")
	apiRouter.HandleFunc("/games/by-name/{name}/actions", api.GameAvailableActionsHandler).Methods("GET")
	apiRouter.HandleFunc("/games/by-name/{name}/execute", api.ExecuteActionHandler).Methods("POST")
	apiRouter.HandleFunc("/games/by-name/{name}/truncate", api.GameTruncateHandler).Methods("POST")
	apiRouter.HandleFunc("/games/by-name/{name}/steps", api.GameByNameStepsHandler).Methods("GET")
	
	// Redirect root to navigation only if no query parameters
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.URL.RawQuery == "" {
			http.Redirect(w, r, "/nav.html", http.StatusFound)
			return
		}
		// If root path has query parameters, serve index.html
		if r.URL.Path == "/" && r.URL.RawQuery != "" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		http.FileServer(http.Dir("./static/")).ServeHTTP(w, r)
	}).Methods("GET")
	
	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	port := "8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Game viewer UI: http://localhost:%s/\n", port)
	fmt.Printf("API endpoints:\n")
	fmt.Printf("  Deck Management:\n")
	fmt.Printf("    - http://localhost:%s/api/decks\n", port)
	fmt.Printf("  Game Management:\n")
	fmt.Printf("    - http://localhost:%s/api/games\n", port)
	
	log.Fatal(http.ListenAndServe(":"+port, r))
}