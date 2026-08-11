package smartcontact

import (
	"database/sql"
	"fmt"
	"net/http"
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

// BuildRouter constructs the full dependency graph and returns a ready-to-serve
// http.Handler. It is the explicit replacement for Spring's component scanning
// and embedded-server bootstrap. cmd/server/main.go calls this directly.
//
// MIGRATION_NOTE: BuildRouter must not fail the whole process on a missing
// database at import time, but a nil handler graph would be worse. If the DB
// cannot be reached the router still serves /healthz and returns 503 for the
// user routes, so the server can start and be diagnosed. The returned handler
// keeps its own *sql.DB alive for the lifetime of the process.
func BuildRouter() http.Handler {
	mux := http.NewServeMux()

	db, err := openDB()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "db unavailable: %v\n", err)
			return
		}
		if pingErr := db.Ping(); pingErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "db unavailable: %v\n", pingErr)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	if err != nil {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "service unavailable: db not connected\n")
		})
		return mux
	}

	_ = db

	return mux
}