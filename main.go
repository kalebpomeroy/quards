package main

import (
	"log"
	
	"github.com/joho/godotenv"
	app "quards/internal/api"
)

func main() {
	// Load .env file (ignore error if file doesn't exist, for production deployments)
	_ = godotenv.Load()
	
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
