package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig returns default test database configuration
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "p@ssw0rd",
		DBName:   "tenantly_test",
	}
}

// SetupTestDB creates a test database connection and runs migrations
// Returns the database connection and a cleanup function
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	config := DefaultTestDBConfig()
	return SetupTestDBWithConfig(t, config)
}

// SetupTestDBWithConfig allows custom configuration
func SetupTestDBWithConfig(t *testing.T, config *TestDBConfig) (*sql.DB, func()) {
	// Build connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
		return nil, nil
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Skipping test: Cannot connect to PostgreSQL: %v", err)
		return nil, nil
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		// Drop all tables (run down migrations)
		if err := dropAllTables(db); err != nil {
			t.Logf("Warning: Failed to cleanup database: %v", err)
		}
		db.Close()
	}

	return db, cleanup
}

// runMigrations runs all up migrations
func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations", // Relative path from test files
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// dropAllTables runs all down migrations to clean up
func dropAllTables(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// CreateTestProperty is a helper to create a test property for repository tests
func CreateTestProperty(t *testing.T, db *sql.DB) int {
	var propertyID int
	err := db.QueryRow(`
		INSERT INTO properties (property_name, property_code, address, property_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		"Test Property", "TEST001", "123 Test Street", "Commercial").Scan(&propertyID)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}
	return propertyID
}
