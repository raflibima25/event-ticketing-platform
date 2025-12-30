package utility

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all SQL migration files
func RunMigrations(db *sql.DB, migrationsPath string) error {
	log.Println("🔄 Running database migrations...")

	// Create migrations table if not exists
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort files to ensure correct order
	sort.Strings(files)

	if len(files) == 0 {
		log.Println("⚠️  No migration files found")
		return nil
	}

	// Execute each migration
	for _, file := range files {
		// Get migration version from filename
		version := filepath.Base(file)
		version = strings.TrimSuffix(version, ".up.sql")

		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("⏭️  Skipping migration %s (already applied)", version)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration
		log.Printf("▶️  Applying migration: %s", version)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		// Record migration
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		log.Printf("✅ Applied migration: %s", version)
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}
