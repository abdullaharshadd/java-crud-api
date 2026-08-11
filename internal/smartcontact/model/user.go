package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// User represents an application user, migrated from the JPA entity
// com.smartContact.model.User mapped to the "user" table.
//
// MIGRATION_NOTE: The Java entity used the table name "USER" and
// mixed-case column names (User_id, User_name, User_Email, ...). Since
// USER is a reserved word in PostgreSQL and the target dialect prefers
// lower_snake_case identifiers to avoid quoting, the table is named
// "users" and columns are lower_snake_case. The field set (id, name,
// email, password, role, about) is preserved exactly.
//
// MIGRATION_NOTE: Name, Email, Password, Role and About are modeled as
// *string so an omitted field marshals/persists as SQL NULL, mirroring
// JPA's merge semantics where a null property leaves/sets the column to
// NULL. A nil pointer encodes to JSON null via the standard library.
type User struct {
	// ID is the auto-generated primary key (Java: User_id, GenerationType.AUTO).
	ID int64 `json:"id"`
	// Name is the user's name (Java: User_name, @NotBlank).
	Name *string `json:"name"`
	// Email is the user's unique email address (Java: User_Email, unique).
	Email *string `json:"email"`
	// Password is the user's password (Java: User_Password).
	Password *string `json:"password"`
	// Role is the user's role (Java: User_Role).
	Role *string `json:"role"`
	// About is a free-text description, up to 500 characters (Java: User_About, length=500).
	About *string `json:"about"`
}

// NewUser constructs a User with the given fields. It mirrors the Java
// @AllArgsConstructor/@Builder while keeping Go's zero-value friendliness.
func NewUser(name, email, password, role, about *string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate enforces the Bean Validation constraints declared on the source
// entity. Currently only Name carries @NotBlank.
//
// MIGRATION_NOTE: Java's @NotBlank rejects null, empty and whitespace-only
// values. The same rule is applied here; the original message is preserved.
func (u *User) Validate() error {
	if u.Name == nil || strings.TrimSpace(*u.Name) == "" {
		return fmt.Errorf("please Add the department Name")
	}
	return nil
}

// CreateUserTableSQL is the DDL that creates the users table. It is derived
// directly from the source JPA entity's fields and replaces Hibernate's
// ddl-auto schema generation.
//
// MIGRATION_NOTE: id uses GENERATED ALWAYS AS IDENTITY (Postgres) in place
// of JPA's GenerationType.AUTO. The email uniqueness constraint from
// @Column(unique = true) and the about length limit from length=500 are
// preserved.
const CreateUserTableSQL = `CREATE TABLE IF NOT EXISTS users (
	id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name     TEXT,
	email    TEXT UNIQUE,
	password TEXT,
	role     TEXT,
	about    VARCHAR(500)
)`

// EnsureUserSchema creates the users table if it does not already exist.
// It should be invoked at application startup so the schema is materialized
// from the model itself, matching the source's ORM-driven schema creation.
func EnsureUserSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, CreateUserTableSQL); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}
	return nil
}
