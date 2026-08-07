package model

import "errors"

// ErrNameRequired is returned by Validate when the Name field is blank.
var ErrNameRequired = errors.New("name is required")

// User represents an application user stored in the users table.
//
// It corresponds to the original JPA User entity. The struct tags map the
// fields to their PostgreSQL column names and to their JSON representation.
//
// MIGRATION_NOTE: The source used mixed-case, quoted-style column names
// (User_id, User_Email, ...). Per the PostgreSQL target dialect guidance we
// prefer lower_snake_case column names so identifiers do not require
// double-quoting. The db tags below therefore use snake_case. If the existing
// database schema literally uses the mixed-case names, either rename the
// columns via migration or adjust the db tags (and quote them in queries).
type User struct {
	// ID is the auto-generated primary key. It is populated by the database
	// via a RETURNING clause on INSERT, so it must not be discarded after save.
	ID int64 `db:"id" json:"id"`

	// Name is the user's display name. It must not be blank.
	Name string `db:"name" json:"name"`

	// Email is the user's email address and is unique across all users.
	Email string `db:"email" json:"email"`

	// Password is the user's (hashed) password.
	Password string `db:"password" json:"password"`

	// Role describes the user's authorization role.
	Role string `db:"role" json:"role"`

	// About is an optional free-text description (max 500 chars in source schema).
	About string `db:"about" json:"about"`
}

// NewUser constructs a User value. The ID is intentionally left as its zero
// value; it is assigned by the database when the record is inserted.
func NewUser(name, email, password, role, about string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate checks the required-field constraints originally expressed via
// Bean Validation (@NotBlank). It returns an error describing the first
// violation found, or nil if the User is valid.
func (u *User) Validate() error {
	if u.Name == "" {
		return ErrNameRequired
	}
	return nil
}