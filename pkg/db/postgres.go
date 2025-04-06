// pkg/db/postgres.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"concert-ticket-api/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.Database) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// WaitForDatabase waits for the database to be available with a timeout
func WaitForDatabase(cfg config.Database, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", cfg.DSN())
		if err == nil {
			err = db.Ping()
			db.Close()

			if err == nil {
				// Database is available
				return nil
			}
		}

		// Wait and retry
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for database after %s", timeout)
}

// RunMigrations runs database migrations
func RunMigrations(cfg config.Database, migrationsPath string) error {
	migrationDSN := cfg.MigrationDSN()
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), migrationDSN)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CreateDatabase creates the database if it doesn't exist
// This is useful for development and test environments
func CreateDatabase(cfg config.Database) error {
	// Create a copy of the config with the 'postgres' database
	tempCfg := cfg
	tempCfg.Name = "postgres"

	// Connect to the 'postgres' database
	db, err := sqlx.Connect("postgres", tempCfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if the database exists
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Name)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create the database if it doesn't exist
	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Name))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	return nil
}
