package testutil

import (
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
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
func SetupTestDB(t *testing.T) (*sqlx.DB, func()) {
	config := DefaultTestDBConfig()
	return SetupTestDBWithConfig(t, config)
}

// SetupTestDBWithConfig allows custom configuration
func SetupTestDBWithConfig(t *testing.T, config *TestDBConfig) (*sqlx.DB, func()) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test: Cannot connect to PostgreSQL: %v", err)
		return nil, nil
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		if err := dropAllTables(db); err != nil {
			t.Logf("Warning: Failed to cleanup database: %v", err)
		}
		db.Close()
	}

	return db, cleanup
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
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

func dropAllTables(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
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
func CreateTestProperty(t *testing.T, db *sqlx.DB) int {
	var propertyID int
	err := db.QueryRow(`
		INSERT INTO properties (property_name, property_code, address, property_type, organization_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		"Test Property", "TEST001", "123 Test Street", "Commercial", 1).Scan(&propertyID)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}
	return propertyID
}

// CreateTestBuilding is a helper to create a test building for repository tests
func CreateTestBuilding(t *testing.T, db *sqlx.DB, propertyID int, orgID int) int {
	var buildingID int
	err := db.QueryRow(`
		INSERT INTO buildings (property_id, building_name, building_code, building_type, organization_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		propertyID, "Test Building", "BLD001", "Office", orgID).Scan(&buildingID)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	return buildingID
}

// CreateTestUnit is a helper to create a test unit for repository tests
func CreateTestUnit(t *testing.T, db *sqlx.DB, buildingID int, orgID int) int {
	var unitID int
	err := db.QueryRow(`
		INSERT INTO units (building_id, unit_number, unit_type, floor_number, area_sqft, organization_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		buildingID, "UNIT001", "Office Space", 1, 500, orgID).Scan(&unitID)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}
	return unitID
}

// CreateTestTenant is a helper to create a test tenant for repository tests
func CreateTestTenant(t *testing.T, db *sqlx.DB, orgID int) int {
	var tenantID int
	err := db.QueryRow(`
		INSERT INTO tenants (name, tenant_type, phone_number, email, nid_last_four, address, organization_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		"Test Tenant", "Individual", "+8801234567890", "test@example.com", "7890", "123 Test Street", orgID).Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// CreateTestUser is a helper to create a test user for repository tests
func CreateTestUser(t *testing.T, db *sqlx.DB) int {
	var userID int
	err := db.QueryRow(`
		INSERT INTO users (username, email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		"testuser", "test@example.com", "hashedpassword", "Test User", "Admin").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// CreateTestOrganization is a helper to create a test organization for repository tests
func CreateTestOrganization(t *testing.T, db *sqlx.DB) int {
	var orgID int
	err := db.QueryRow(`
		INSERT INTO organizations (org_name, org_code, org_type, contact_email)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		"Test Organization", "TEST001", "PropertyManagement", "org@example.com").Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}
	return orgID
}
