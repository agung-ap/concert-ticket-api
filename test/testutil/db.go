package testutil

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	testDB     *sqlx.DB
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	dbName     = "test_db"
	dbUser     = "postgres"
	dbPassword = "postgres"
	dbPort     string
)

// SetupTestDB creates a test database in Docker
func SetupTestDB() (*sqlx.DB, error) {
	// Skip Docker if we already have a test DB connection
	if testDB != nil {
		return testDB, nil
	}

	// Create a new Docker pool
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Start PostgreSQL container
	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// Set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %w", err)
	}

	// Get the container port
	dbPort = resource.GetPort("5432/tcp")

	// Set a 30 second timeout
	if err = pool.Retry(func() error {
		// Try to connect to the database
		testDB, err = sqlx.Connect("postgres", fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", dbUser, dbPassword, dbPort, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		// If it fails, kill the container
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			fmt.Printf("Could not purge resource: %s\n", purgeErr)
		}
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Create the tables
	err = createTables(testDB)
	if err != nil {
		return nil, fmt.Errorf("could not create tables: %w", err)
	}

	return testDB, nil
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(db *sqlx.DB) error {
	// Truncate all tables
	_, err := db.Exec("TRUNCATE TABLE bookings, concerts RESTART IDENTITY CASCADE")
	return err
}

// TeardownTestDB tears down the test database
func TeardownTestDB() {
	if testDB != nil {
		testDB.Close()
	}

	if resource != nil && pool != nil {
		if err := pool.Purge(resource); err != nil {
			fmt.Printf("Could not purge resource: %s\n", err)
		}
	}
}

// createTables creates the necessary tables for testing
func createTables(db *sqlx.DB) error {
	// Create concerts table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS concerts (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			artist VARCHAR(255) NOT NULL,
			venue VARCHAR(255) NOT NULL,
			concert_date TIMESTAMP NOT NULL,
			total_tickets INT NOT NULL,
			available_tickets INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			booking_start_time TIMESTAMP NOT NULL,
			booking_end_time TIMESTAMP NOT NULL,
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create bookings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookings (
			id SERIAL PRIMARY KEY,
			concert_id INT NOT NULL REFERENCES concerts(id),
			user_id VARCHAR(255) NOT NULL,
			ticket_count INT NOT NULL,
			booking_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT valid_ticket_count CHECK (ticket_count > 0)
		)
	`)
	return err
}
