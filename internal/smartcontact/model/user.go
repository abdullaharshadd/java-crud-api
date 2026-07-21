// Package model contains the domain entities and data transfer objects
// (DTOs) for the SmartContact service. It is the Go equivalent of the Java
// package com.smartContact.model.
//
// MIGRATION_NOTE: The source User.java was a JPA @Entity mapped to the
// "USER" table, with Lombok generating getters, setters, equals, hashCode,
// toString, a no-arg constructor, an all-args constructor, and a builder.
// Bean Validation (@NotBlank) enforced a non-empty name.
//
// In idiomatic Go these concerns are handled explicitly and separately:
//
//   - The Lombok-generated boilerplate collapses into a plain struct with
//     exported fields. Field access replaces getters/setters, so no
//     accessor methods are required.
//   - JPA/Hibernate ORM mapping (@Entity, @Table, @Column, @Id,
//     @GeneratedValue) has no direct Go equivalent. There is no
//     reflection-based ORM in the standard library. The original column
//     names are preserved in `db` struct tags so a data-access layer using
//     sqlx/database/sql can map rows explicitly. The auto-generated primary
//     key becomes the responsibility of the database (e.g. SERIAL /
//     AUTO_INCREMENT) and/or the repository layer.
//   - JSON tags are added so the struct can be marshalled directly for API
//     responses. Note: Password is tagged `json:"-"` so it is never leaked
//     in serialized output. Review this if the API contract requires the
//     password field.
//   - Bean Validation (@NotBlank on name) is reimplemented as an explicit
//     Validate method returning an error, since Go has no annotation-driven
//     validation framework in the standard library.
//   - The @Builder pattern is replaced by a NewUser constructor. If optional
//     construction becomes necessary, migrate to functional options.
package model

import (
	"errors"
	"strings"
)

// ErrUserNameBlank is returned by User.Validate when the user's name is
// empty or consists solely of whitespace.
//
// MIGRATION_NOTE: This replaces the Bean Validation constraint
// @NotBlank(message = "please Add the department Name"). The original
// message is preserved verbatim for compatibility with existing clients,
// even though it refers to a "department Name" rather than a user name.
var ErrUserNameBlank = errors.New("please Add the department Name")

// User represents an application user. It is the Go equivalent of the JPA
// entity com.smartContact.model.User, previously mapped to the "USER"
// database table.
//
// The struct tags preserve the original database column names so a
// data-access layer can map query results explicitly. Field access replaces
// the Lombok-generated getters and setters.
type User struct {
	// ID is the primary key. In the source it was an auto-generated column
	// ("User_id") via @GeneratedValue(strategy = AUTO); generation is now the
	// responsibility of the database and/or repository layer.
	ID int `json:"id" db:"User_id"`

	// Name is the user's name. It must not be blank (see Validate).
	// Mapped to the "User_name" column.
	Name string `json:"name" db:"User_name"`

	// Email is the user's unique email address. Mapped to the "User_Email"
	// column, which carried a UNIQUE constraint in the source schema.
	Email string `json:"email" db:"User_Email"`

	// Password is the user's password. Mapped to the "User_Password" column.
	//
	// MIGRATION_NOTE: This field is excluded from JSON output (`json:"-"`) to
	// avoid leaking credentials in API responses. Adjust the tag if the
	// original API contract intentionally serialized this field.
	Password string `json:"-" db:"User_Password"`

	// Role is the user's role. Mapped to the "User_Role" column.
	Role string `json:"role" db:"User_Role"`

	// About is a free-form description of the user, limited to 500
	// characters in the source schema. Mapped to the "User_About" column.
	About string `json:"about" db:"User_About"`
}

// NewUser constructs a User with all fields populated. It is the idiomatic
// replacement for the Lombok @AllArgsConstructor and @Builder. The zero
// value of User (an empty struct) replaces the @NoArgsConstructor.
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

// Validate checks the domain invariants for a User. It returns
// ErrUserNameBlank if the Name field is empty or whitespace-only,
// mirroring the source @NotBlank constraint. It returns nil when the User
// is valid.
//
// MIGRATION_NOTE: JPA/Bean Validation applied this automatically on
// persist. In Go the caller must invoke Validate explicitly (typically in
// the service or repository layer before writing to the database).
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrUserNameBlank
	}
	return nil
}
