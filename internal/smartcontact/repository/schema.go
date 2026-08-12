package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// InitSchema creates the users table if it does not already exist.
// Column names match the db struct tags on model.User exactly:
// id, name, email, password, role, about.
func InitSchema(ctx context.Context, db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    email    TEXT,
    password TEXT,
    role     TEXT,
    about    TEXT
);`
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}