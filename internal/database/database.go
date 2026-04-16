package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// InitDB initializes database connection
func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite is single-writer
	db.SetMaxIdleConns(1)

	return db, nil
}
