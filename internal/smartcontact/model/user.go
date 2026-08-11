package model

// This file adds the User domain entity to the model package. See
// errormessage.go for the package-level doc comment.
//
// MIGRATION_NOTE: The Java source declared User as a JPA @Entity mapped to the
// "USER" table, with Lombok generating getters/setters/constructor/builder and
// Bean Validation (@NotBlank) on the name field. Go has no ORM annotations, no
// Lombok, and no annotation-driven validation, so:
//
//   - Exported struct fields replace Lombok getters/setters.
//   - JPA @Column names map to `db` struct tags for use with sqlx-style
//     scanning; `json` tags control API serialization. The Password field is
//     tagged `json:"-"` so it is never leaked in API responses.
//   - The @NotBlank validation becomes an explicit Validate() method rather
//     than a magic annotation, returning an error the caller must handle.
//   - A NewUser constructor replaces Lombok's @Builder / @AllArgsConstructor.
//
// DATABASE_NOTE: The target database is PostgreSQL. The original JPA mapping
// used mixed-case, quoted column names (e.g. "User_id", "User_Email"). The Go
// `db` tags below use lower_snake_case (user_id, user_name, ...) so identifiers
// need no quoting; the ID is expected to be backed by a GENERATED ALWAYS AS
// IDENTITY / SERIAL column, and inserts should use `RETURNING user_id` rather
// than any LastInsertId() call. Adjust the tags here to match the actual
// schema if the DDL was migrated verbatim with quoted mixed-case names.

import (
	"errors"
	"strings"
)

// ErrUserNameBlank is returned by User.Validate when the user's name is empty
// or consists solely of whitespace. It replaces the Java @NotBlank constraint
// message "please Add the department Name".
var ErrUserNameBlank = errors.New("user name must not be blank")

// User is the domain entity representing an application user. It is the Go
// equivalent of the JPA-mapped com.smartContact.model.User class.
type User struct {
	// ID is the primary key. In PostgreSQL this is backed by an identity/serial
	// column and is populated via `RETURNING user_id` on insert.
	ID int `db:"user_id" json:"id"`

	// Name is the user's display name. It must not be blank (see Validate).
	Name string `db:"user_name" json:"name"`

	// Email is the user's email address. It is unique at the database level.
	Email string `db:"user_email" json:"email"`

	// Password is the user's (hashed) password. It is excluded from JSON
	// serialization so it is never exposed through the API layer.
	Password string `db:"user_password" json:"-"`

	// Role is the user's authorization role.
	Role string `db:"user_role" json:"role"`

	// About is a free-form description of the user (max 500 characters in the
	// original schema).
	About string `db:"user_about" json:"about"`
}

// NewUser constructs a User from the given field values. It replaces Lombok's
// @AllArgsConstructor / @Builder. The ID is intentionally omitted because it is
// assigned by the database on insert.
func NewUser(name, email, password, role, about string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate enforces the invariants that JPA/Bean Validation checked via
// annotations. It returns ErrUserNameBlank if the name is empty or whitespace,
// mirroring the original @NotBlank constraint. Callers must invoke this
// explicitly before persisting a User.
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrUserNameBlank
	}
	return nil
}
