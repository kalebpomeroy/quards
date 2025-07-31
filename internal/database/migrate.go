package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all pending database migrations
func RunMigrations() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Ensure migrations tracking table exists
	err := createMigrationsTable(db)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Apply pending migrations
	for _, filename := range migrationFiles {
		version := extractVersionFromFilename(filename)
		if _, applied := appliedMigrations[version]; applied {
			fmt.Printf("Migration %s already applied, skipping\n", version)
			continue
		}

		fmt.Printf("Applying migration %s...\n", version)
		err = applyMigration(db, filename, version)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}
		fmt.Printf("Migration %s applied successfully\n", version)
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles returns sorted list of migration files
func getMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files to ensure proper order
	sort.Strings(files)
	return files, nil
}

// extractVersionFromFilename extracts migration version from filename
func extractVersionFromFilename(filename string) string {
	base := filepath.Base(filename)
	// Remove .sql extension
	base = strings.TrimSuffix(base, ".sql")
	return base
}

// applyMigration executes a single migration file
func applyMigration(db *sql.DB, filename, version string) error {
	// Read migration file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Note: This is a simple approach. For complex migrations with multiple statements,
	// you might need to split the SQL content and execute statements separately
	_, err = tx.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied (only if not already recorded in the SQL)
	if !strings.Contains(string(content), "INSERT INTO schema_migrations") {
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}