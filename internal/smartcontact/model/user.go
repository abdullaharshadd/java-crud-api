package model

import (
	"strings"
)

// User is the database representation of an application user, mapping to the
// "users" table. It replaces the Java JPA entity com.smartContact.model.User.
//
// MIGRATION_NOTE: The Java entity relied on JPA annotations (@Entity, @Table,
// @Column, @Id, @GeneratedValue) plus Lombok (@Data, @Builder, etc.) to map to
// the database and generate boilerplate. In idiomatic Go there is no ORM
// annotation layer here; the repository layer is responsible for the SQL
// mapping. Column names are expressed as db tags for sqlx-style scanning and use
// PostgreSQL-friendly lower_snake_case (the original mixed-case "User_id",
// "User_Email", etc. would have required awkward double-quoting on every query).
// The primary key is a database-generated identity column populated via
// INSERT ... RETURNING id in the repository, replacing JPA's
// GenerationType.AUTO.
//
// Per migration deviation 1, the model is split into three shapes: this
// database struct (User), a request DTO (CreateUserRequest) carrying the
// plaintext password on input, and a response DTO (UserResponse) which
// deliberately omits the password field so credentials are never serialized to
// clients.
type User struct {
	// ID is the database-generated primary key (maps to column user_id).
	ID int `db:"user_id"`
	// Name is the user's display name (maps to column user_name). Required.
	Name string `db:"user_name"`
	// Email is the user's unique email address (maps to column user_email).
	Email string `db:"user_email"`
	// Password is the user's (hashed) password (maps to column user_password).
	Password string `db:"user_password"`
	// Role is the user's authorization role (maps to column user_role).
	Role string `db:"user_role"`
	// About is a free-form user description, limited to 500 characters
	// (maps to column user_about).
	About string `db:"user_about"`
}

// CreateUserRequest is the inbound DTO used when creating or updating a user.
// It carries the plaintext password supplied by the client, which the service
// layer is expected to hash before persisting.
type CreateUserRequest struct {
	// Name is the user's display name. Required (see Validate).
	Name string `json:"name"`
	// Email is the user's unique email address.
	Email string `json:"email"`
	// Password is the plaintext password supplied by the client.
	Password string `json:"password"`
	// Role is the user's authorization role.
	Role string `json:"role"`
	// About is a free-form user description.
	About string `json:"about"`
}

// UserResponse is the outbound DTO returned to clients. It intentionally omits
// the Password field so that credentials are never serialized in API responses
// (migration deviation 1).
type UserResponse struct {
	// ID is the database-generated primary key.
	ID int `json:"id"`
	// Name is the user's display name.
	Name string `json:"name"`
	// Email is the user's unique email address.
	Email string `json:"email"`
	// Role is the user's authorization role.
	Role string `json:"role"`
	// About is a free-form user description.
	About string `json:"about"`
}

// maxAboutLength mirrors the JPA @Column(length = 500) constraint on the
// user_about column.
const maxAboutLength = 500

// Validate checks the request against the constraints declared on the original
// JPA entity: the @NotBlank rule on the name field and the length limit on the
// about field. It returns a slice of human-readable validation messages; an
// empty (nil) result means the request is valid.
//
// MIGRATION_NOTE: This replaces Bean Validation's @NotBlank annotation. The
// original message ("please Add the department Name") appears to be a copy/paste
// artifact referring to a department rather than a user; the message here is
// corrected to reference the user's name.
func (r CreateUserRequest) Validate() []string {
	var problems []string
	if strings.TrimSpace(r.Name) == "" {
		problems = append(problems, "please add the user name")
	}
	if len(r.About) > maxAboutLength {
		problems = append(problems, "about must be at most 500 characters")
	}
	return problems
}

// ToUser converts an inbound CreateUserRequest into the database User struct.
// The ID is left zero because it is assigned by the database on insert.
func (r CreateUserRequest) ToUser() User {
	return User{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
		Role:     r.Role,
		About:    r.About,
	}
}

// ToResponse projects a database User into the client-facing UserResponse,
// deliberately dropping the Password field.
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
		About: u.About,
	}
}
