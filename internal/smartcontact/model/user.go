package model

import (
	"strings"
)

// User represents an application user, mapping to the "users" table in the
// database. It corresponds to the original Java JPA entity
// com.smartContact.model.User.
//
// MIGRATION_NOTE: The Java source relied on JPA annotations (@Entity, @Table,
// @Column, @Id, @GeneratedValue) and Lombok (@Data, @NoArgsConstructor,
// @AllArgsConstructor, @Builder) to generate persistence mapping and
// boilerplate. Go has no ORM annotations or code generation for this; instead
// the struct fields are exported directly with `db` tags for sqlx and `json`
// tags for HTTP (de)serialization. Getters/setters are unnecessary in
// idiomatic Go for a plain data struct.
//
// MIGRATION_NOTE: The original table/column names used mixed case
// ("USER", "User_id", "User_Email", ...). Per the target PostgreSQL dialect we
// use lower_snake_case identifiers so no double-quoting is required. If the
// existing database schema truly uses the mixed-case names, the `db` tags and
// UserTable/UserColumns constants below must be adjusted (and identifiers
// double-quoted) to match — MANUAL REVIEW REQUIRED.
//
// MIGRATION_NOTE: The Java @NotBlank(message = "please Add the department Name")
// validation on `name` is enforced here by the Validate method rather than a
// framework annotation, since Go has no equivalent bean-validation runtime.
type User struct {
	// ID is the auto-generated primary key (original @Id @GeneratedValue AUTO,
	// column "User_id"). In PostgreSQL this is populated via a RETURNING id
	// clause on INSERT rather than an auto-increment callback.
	ID int `db:"id" json:"id"`

	// Name is the user's display name (original column "User_name").
	// It must not be blank; see Validate.
	Name string `db:"name" json:"name"`

	// Email is the user's email address (original column "User_Email").
	// It is subject to a unique constraint at the database level.
	Email string `db:"email" json:"email"`

	// Password is the user's (hashed) password (original column "User_Password").
	Password string `db:"password" json:"password"`

	// Role is the user's role (original column "User_Role").
	Role string `db:"role" json:"role"`

	// About is a free-text description of the user, up to 500 characters
	// (original column "User_About", length = 500).
	About string `db:"about" json:"about"`
}

const (
	// UserTable is the name of the database table backing User.
	UserTable = "users"

	// UserColumns is the comma-separated list of database columns for User,
	// in struct-field order. It is intended for building SQL statements.
	UserColumns = "id, name, email, password, role, about"

	// UserAboutMaxLength is the maximum allowed length of the About field,
	// mirroring the original @Column(length = 500) constraint.
	UserAboutMaxLength = 500

	// errNameBlank is the validation message from the original
	// @NotBlank annotation on the name field.
	errNameBlank = "please Add the department Name"
)

// NewUser constructs a User with the given fields. It corresponds to the
// Lombok-generated all-args constructor. Callers that need to persist a new
// user typically leave id as 0 and let the database assign it.
func NewUser(id int, name, email, password, role, about string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate reports whether the User satisfies the constraints carried over from
// the Java bean-validation annotations. It returns an error describing the
// first violation, or nil if the User is valid.
//
// MIGRATION_NOTE: In Java these checks were enforced automatically by the
// validation framework at the web/persistence boundary. In Go validation is
// explicit; callers must invoke Validate at the appropriate point (typically in
// the service or handler layer before persisting).
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return &ValidationError{Field: "name", Message: errNameBlank}
	}
	if len(u.About) > UserAboutMaxLength {
		return &ValidationError{Field: "about", Message: "about must be at most 500 characters"}
	}
	return nil
}

// ValidationError describes a single field validation failure on a model type.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string
	// Message is a human-readable description of the failure.
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
