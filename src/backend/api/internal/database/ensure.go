package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/lib/pq"
)

// EnsureDatabaseExists creates the database named in databaseURL if it
// doesn't already exist yet, connecting to the server's default "postgres"
// maintenance database to do so. This is a local-development convenience
// only (see the call site in cmd/server/main.go for the production gating):
// it lets a contributor point DATABASE_URL at a bare Postgres server without
// first running `createdb` by hand.
func EnsureDatabaseExists(databaseURL string) error {
	name, adminURL, err := adminConnectionURL(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	adminDB, err := sql.Open("postgres", adminURL)
	if err != nil {
		return fmt.Errorf("failed to open maintenance connection: %w", err)
	}
	defer func() { _ = adminDB.Close() }()

	var exists bool
	if err := adminDB.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name,
	).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if database %q exists: %w", name, err)
	}
	if exists {
		return nil
	}

	if _, err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(name))); err != nil {
		return fmt.Errorf("failed to create database %q: %w", name, err)
	}

	return nil
}

// adminConnectionURL extracts the target database name from a postgres
// connection URL and returns a sibling URL pointing at the server's default
// "postgres" maintenance database instead, since a database that doesn't
// exist yet can't be connected to directly.
func adminConnectionURL(databaseURL string) (name string, adminURL string, err error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", "", err
	}

	name = strings.TrimPrefix(u.Path, "/")
	if name == "" {
		return "", "", fmt.Errorf("no database name in URL path")
	}

	admin := *u
	admin.Path = "/postgres"
	return name, admin.String(), nil
}
