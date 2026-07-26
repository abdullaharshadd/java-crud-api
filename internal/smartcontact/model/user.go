// Package model defines the domain types for the SmartContact application.
//
// MIGRATION_NOTE: The original Java type was `com.smartContact.model.User`, a
// JPA @Entity mapped to the USER table and decorated with Lombok annotations
// (@Data, @NoArgsConstructor, @AllArgsConstructor, @Builder) plus a Bean
// Validation @NotBlank constraint on the name field.
//
// Go has no ORM annotations, no Lombok code generation, and no field-level
// validation framework baked into the language. The idiomatic replacements
// used here are:
//   - A plain struct with `db` tags for sqlx column mapping.
//   - A constructor function (NewUser) instead of @Builder / @AllArgsConstructor.
//   - An explicit Validate method instead of the @NotBlank annotation.
//   - Getters/setters are omitted entirely; exported fields are idiomatic Go.
//
// IMPORTANT (sqlx mapping): The original columns use mixed case such as
// "User_name" and "User_Email". sqlx maps result-set column names to struct
// `db` tags case-sensitively for some drivers, so every SELECT that scans into
// User MUST alias the columns to the exact lowercase tag names below, e.g.:
//
//	SELECT User_id   AS user_id,
//	       User_name AS user_name,
//	       User_Email AS user_email,
//	       User_Password AS user_password,
//	       User_Role AS user_role,
//	       User_About AS user_about
//	FROM USER
package model

import (
	"errors"
	"strings"
)

// ErrUserNameBlank is returned by User.Validate when the user's name is empty
// or consists solely of whitespace.
//
// MIGRATION_NOTE: This replaces the Java Bean Validation constraint
// `@NotBlank(message = "please Add the department Name")`. The original
// (mildly incorrect) message referenced a "department Name"; it is preserved
// here to keep behavioral parity for any tests or logs that assert on it.
var ErrUserNameBlank = errors.New("please Add the department Name")

// User represents an application user, mapped to the USER table.
//
// The `db` struct tags correspond to the aliased, lowercase column names that
// callers must produce in their SELECT statements (see the package-level
// MIGRATION_NOTE for details). The Email column carried a UNIQUE constraint in
// the original schema; that constraint lives in the database DDL, not in this
// struct.
type User struct {
	// ID is the auto-generated primary key (JPA GenerationType.AUTO).
	ID int `db:"user_id"`
	// Name is the user's display name. It must not be blank (see Validate).
	Name string `db:"user_name"`
	// Email is the user's email address; unique across all users.
	Email string `db:"user_email"`
	// Password is the user's (hashed) password.
	Password string `db:"user_password"`
	// Role is the user's authorization role.
	Role string `db:"user_role"`
	// About is a free-text description, originally capped at 500 characters.
	About string `db:"user_about"`
}

// AboutMaxLength is the maximum allowed length of the About field, preserved
// from the original JPA @Column(length = 500) definition.
//
// MIGRATION_NOTE: In JPA the length attribute primarily influenced DDL
// generation; there was no runtime enforcement on the Java side. It is exposed
// here as a constant so callers (or the schema) can enforce it explicitly.
const AboutMaxLength = 500

// NewUser constructs a User with the provided fields. It replaces the Lombok
// @Builder / @AllArgsConstructor pair from the original entity.
//
// The ID is intentionally not a parameter: it is assigned by the database on
// insert (auto-increment), mirroring GenerationType.AUTO. Callers that need to
// set an existing ID (for example when loading a record) may assign the ID
// field directly.
func NewUser(name, email, password, role, about string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate checks the User's invariants. It returns ErrUserNameBlank if the
// Name field is empty or whitespace-only, mirroring the original
// @NotBlank constraint. It returns nil when the user is valid.
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrUserNameBlank
	}
	return nil
}
