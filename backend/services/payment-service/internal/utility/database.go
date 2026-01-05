package utility

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
)

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*sql.DB, error) {
	var dsn string

	// Check if using Cloud SQL Unix socket (path starts with /)
	if len(cfg.Host) > 0 && cfg.Host[0] == '/' {
		// Unix socket format: no port needed
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.User,
			cfg.Password,
			cfg.DBName,
			cfg.SSLMode,
		)
	} else {
		// TCP connection format
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.DBName,
			cfg.SSLMode,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
