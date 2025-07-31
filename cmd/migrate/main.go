package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"quards/internal/database"
)

func main() {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	var help bool
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.BoolVar(&help, "h", false, "Show help message (shorthand)")
	flag.Parse()

	if help {
		showHelp()
		return
	}

	// Initialize database connection
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Run migrations
	fmt.Println("Starting database migrations...")
	err = database.RunMigrations()
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Database migrations completed successfully!")
}

func showHelp() {
	fmt.Printf(`Database Migration Tool for Quards

Usage: %s [options]

Options:
  -h, -help     Show this help message

Environment Variables:
  DATABASE_URL  Database connection string (required)
                Example: postgres://user:pass@localhost/dbname?sslmode=disable

Examples:
  # Run migrations with default .env file
  %s

  # Run migrations with custom database URL
  DATABASE_URL="postgres://user:pass@host/db?sslmode=require" %s

  # Run migrations in production
  export DATABASE_URL="your_production_db_url"
  %s

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}