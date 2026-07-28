package model

import "database/sql"

// User represents an application user as stored in the "users" database table.
// It carries authentication and profile fields.
//
// MIGRATION_NOTE: The Java source was a JPA @Entity with Lombok
// (@Data/@Builder/@AllArgsConstructor). In Go we model this as a plain write
// shape. This struct is INTERNAL ONLY — it is never serialized directly to
// clients. Use UserResponse (below) as the single shared wire shape for every
// user-returning endpoint.
//
// MIGRATION_NOTE: Table/column names were normalized to PostgreSQL-friendly
// lower_snake_case (e.g. "USER"."User_id" -> users.id), so no identifier
// quoting is required. The auto-generated primary key (JPA GenerationType.AUTO)
// maps to a PostgreSQL identity/serial column; inserts should use
// "RETURNING id" rather than any LastInsertId() call.
//
// MIGRATION_NOTE: The @NotBlank validation on Name ("please Add the department
// Name") is not a struct tag here; validation should be enforced in the
// service/handler layer before persisting.
type User struct {
	// ID is the primary key (maps to users.id, auto-generated identity column).
	ID int
	// Name is the user's display name. Must not be blank when persisted.
	Name string
	// Email is the user's email address. Unique across the users table.
	Email string
	// Password is the user's (hashed) password.
	Password string
	// Role is the user's authorization role.
	Role string
	// About is a free-text profile field (max length 500 in the schema).
	About string
}

// NewUser constructs a User with the given fields. This replaces Lombok's
// @AllArgsConstructor / @Builder. The ID is typically assigned by the database
// on insert, so callers usually pass 0 (or the zero value) here.
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

// userRow is the scan target for reading a user from the database. Nullable
// columns use sql.NullString so that NULL values are distinguishable from
// empty strings.
//
// MIGRATION_NOTE: This is the DB-facing read shape. Repositories should scan
// into a userRow and then call toResponse to obtain the client-facing shape.
type userRow struct {
	ID       int
	Name     sql.NullString
	Email    sql.NullString
	Password sql.NullString
	Role     sql.NullString
	About    sql.NullString
}

// UserResponse is the client-facing wire shape for any endpoint that returns a
// user. Nullable columns are represented as *string so that JSON omits or
// nulls them appropriately. The password is intentionally never included.
type UserResponse struct {
	// ID is the user's primary key.
	ID int `json:"id"`
	// Name is the user's display name, nil when the column is NULL.
	Name *string `json:"name"`
	// Email is the user's email address, nil when the column is NULL.
	Email *string `json:"email"`
	// Role is the user's authorization role, nil when the column is NULL.
	Role *string `json:"role"`
	// About is the user's profile text, nil when the column is NULL.
	About *string `json:"about"`
}

// nsToPtr converts a sql.NullString into a *string, returning nil when the
// underlying value is NULL.
func nsToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

// toResponse converts a database userRow into the client-facing UserResponse.
// The password field is deliberately dropped and never exposed to clients.
func (r userRow) toResponse() UserResponse {
	return UserResponse{
		ID:    r.ID,
		Name:  nsToPtr(r.Name),
		Email: nsToPtr(r.Email),
		Role:  nsToPtr(r.Role),
		About: nsToPtr(r.About),
	}
}
