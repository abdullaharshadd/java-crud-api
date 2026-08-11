package smartcontact

import (
	"database/sql"
	"fmt"
	"os"
)

// DefaultDatabaseURL is the DSN used when DATABASE_URL is unset.
// MIGRATION_NOTE: In the Java app this came from application.properties;
// here it is read from the environment with a sane local default.
const DefaultDatabaseURL = "postgres://app:app@db:5432/app?sslmode=disable"

// openDB opens the PostgreSQL connection pool and verifies connectivity.
// The caller owns the returned *sql.DB and is responsible for closing it.
func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = DefaultDatabaseURL
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}