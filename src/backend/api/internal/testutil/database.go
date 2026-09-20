package testutil

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

// URL renders the config as a libpq connection string.
func (c *TestDBConfig) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// DefaultTestDatabaseURL returns the connection string every DB-backed test
// should use. TEST_DATABASE_URL overrides it; DATABASE_URL deliberately does
// not, because these suites run migrations and truncate tables — falling back
// to the app's own connection string would point them at the dev database.
func DefaultTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return DefaultTestDBConfig().URL()
}

// SetupTestDB creates a test database connection and runs migrations
// Returns the database connection and a cleanup function
func SetupTestDB(t *testing.T) (*sqlx.DB, func()) {
	config := DefaultTestDBConfig()
	return SetupTestDBWithConfig(t, config)
}

// SetupTestDBNamed is SetupTestDB against a database of its own, created on
// first use. `go test ./internal/...` runs packages in parallel, and cleanup
// here drops every table — so two packages sharing one database will tear down
// each other's schema mid-run. Each package that needs a live database should
// therefore claim its own name.
func SetupTestDBNamed(t *testing.T, dbName string) (*sqlx.DB, func()) {
	if _, err := EnsureTestDatabase(dbName); err != nil {
		t.Skipf("Skipping test: Cannot reach test database %q: %v", dbName, err)
		return nil, nil
	}
	config := DefaultTestDBConfig()
	config.DBName = dbName

	// Reset first: a suite sharing this database may have left rows behind
	// (soft deletes in particular survive their own teardown), and those
	// collide with fixture codes here.
	if err := ResetSchema(config.URL()); err != nil {
		t.Skipf("Skipping test: Cannot reset test database %q: %v", dbName, err)
		return nil, nil
	}

	return SetupTestDBWithConfig(t, config)
}

// EnsureTestDatabase creates dbName if it does not exist yet and returns its
// DSN. TEST_DATABASE_URL overrides both — it names a specific database, so the
// caller's name is ignored and nothing is created.
func EnsureTestDatabase(dbName string) (string, error) {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url, nil
	}

	config := DefaultTestDBConfig()
	config.DBName = dbName
	if err := ensureDatabase(config); err != nil {
		return "", err
	}
	return config.URL(), nil
}

// ensureDatabase creates config.DBName if it does not exist yet, connecting to
// the server's default maintenance database to do so. CREATE DATABASE cannot
// run inside a transaction or be written as IF NOT EXISTS, hence the explicit
// catalog check.
func ensureDatabase(config *TestDBConfig) error {
	maintenance := *config
	maintenance.DBName = "postgres"

	db, err := sqlx.Connect("postgres", maintenance.URL())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	var exists bool
	if err := db.Get(&exists, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", config.DBName); err != nil {
		return err
	}
	if exists {
		return nil
	}

	// pq has no placeholder support for identifiers; the name is ours, not
	// user input, but quote it so an unusual name still parses.
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(config.DBName))); err != nil {
		// A parallel test binary may have won the race between the check and
		// the create; that is not an error for our purposes.
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P04" {
			return nil
		}
		return err
	}

	return nil
}

// RunMigrations applies every migration to the database behind connStr. It
// exists so tests outside this package can migrate a database they opened
// themselves: database.RunMigrations resolves "file://migrations" against the
// process working directory, which is the package directory under `go test`,
// not the module root.
func RunMigrations(connStr string) error {
	return runMigrations(connStr)
}

// ResetSchema drops every table and re-applies the migrations, so a suite
// starts from a known-empty schema. Suites that seed fixed data (a fixed
// organization slug, username, or property code) otherwise fail on the second
// run against the same database — including after a run that aborted before
// its teardown.
func ResetSchema(connStr string) error {
	if err := dropAllTables(connStr); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	return runMigrations(connStr)
}

// SetupTestDBWithConfig allows custom configuration
func SetupTestDBWithConfig(t *testing.T, config *TestDBConfig) (*sqlx.DB, func()) {
	connStr := config.URL()

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test: Cannot connect to PostgreSQL: %v", err)
		return nil, nil
	}

	if err := runMigrations(connStr); err != nil {
		_ = db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		if err := dropAllTables(connStr); err != nil {
			t.Logf("Warning: Failed to cleanup database: %v", err)
		}
		_ = db.Close()
	}

	return db, cleanup
}

// runMigrations and dropAllTables take a connection string rather than the
// shared *sqlx.DB on purpose: migrate.New opens and owns its own independent
// *sql.DB internally, so its Close() only ever closes that. Passing the
// shared connection through postgres.WithInstance instead (the previous
// approach) hands migrate.Migrate.Close() a reference to OUR db handle —
// its postgres driver's Close() unconditionally closes whatever *sql.DB it
// was given, which killed the connection every test needed for the rest of
// its body immediately after setup ("sql: database is closed").
func runMigrations(connStr string) error {
	m, err := newMigrate(connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func dropAllTables(connStr string) error {
	m, err := newMigrate(connStr)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// newMigrate opens a migrate instance and clears any dirty flag first.
//
// A migration that fails partway leaves the version marked dirty, and every
// later run then refuses to do anything until someone forces the version by
// hand — one bad run poisons the database for the rest of the day. That is
// intolerable for a disposable test database, where the schema is rebuilt from
// scratch constantly, so recover automatically instead. Only ever call this
// against a test database.
func newMigrate(connStr string) (*migrate.Migrate, error) {
	m, err := migrate.New("file://../../migrations", connStr)
	if err != nil {
		return nil, err
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		_, _ = m.Close()
		return nil, err
	}
	if dirty {
		if err := m.Force(int(version)); err != nil {
			_, _ = m.Close()
			return nil, fmt.Errorf("failed to clear dirty version %d: %w", version, err)
		}
	}

	return m, nil
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
	// building_code is unique per property (buildings_property_id_building_code_key),
	// and some tests create more than one building on the same property.
	n := testEntityCounter.Add(1)

	var buildingID int
	err := db.QueryRow(`
		INSERT INTO buildings (property_id, building_name, building_code, building_type, organization_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		propertyID, "Test Building", fmt.Sprintf("BLD%03d", n), "Commercial", orgID).Scan(&buildingID)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	return buildingID
}

// CreateTestUnit is a helper to create a test unit for repository tests.
// property_id is derived from the building rather than taken as a param —
// a unit's property must always match its building's, so this can't drift
// out of sync the way a separately-passed value could.
func CreateTestUnit(t *testing.T, db *sqlx.DB, buildingID int, orgID int) int {
	var propertyID int
	if err := db.QueryRow(`SELECT property_id FROM buildings WHERE id = $1`, buildingID).Scan(&propertyID); err != nil {
		t.Fatalf("Failed to look up property for test building: %v", err)
	}

	// unit_number is unique per building (units_building_id_unit_number_key),
	// and some tests create several units on the same building.
	n := testEntityCounter.Add(1)

	var unitID int
	err := db.QueryRow(`
		INSERT INTO units (building_id, property_id, unit_number, unit_type, floor, organization_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		buildingID, propertyID, fmt.Sprintf("UNIT%03d", n), "Office", 1, orgID).Scan(&unitID)
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
	n := testEntityCounter.Add(1)
	err := db.QueryRow(`
		INSERT INTO users (username, email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		fmt.Sprintf("testuser%d", n), fmt.Sprintf("test%d@example.com", n), "hashedpassword", "Test", "User", "Admin").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// testEntityCounter gives each test-created row that needs a unique value
// (e.g. a slug or email with a UNIQUE constraint) a distinct suffix — some
// tests create more than one of these within the same database.
var testEntityCounter atomic.Int64

// CreateTestOrganization is a helper to create a test organization for repository tests
func CreateTestOrganization(t *testing.T, db *sqlx.DB) int {
	var orgID int
	n := testEntityCounter.Add(1)
	err := db.QueryRow(`
		INSERT INTO organizations (name, slug, subscription_tier)
		VALUES ($1, $2, $3)
		RETURNING id`,
		"Test Organization", fmt.Sprintf("test-org-%d", n), "basic").Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}
	return orgID
}
