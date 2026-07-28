package model

// User is the domain model representing an application user.
//
// In the original Java source this was a JPA @Entity (Lombok-generated
// getters/setters/builder) mapped to the "USER" table. Here it is a plain Go
// struct: persistence concerns (column mapping, unique constraints, generated
// keys) live in the repository/SQL layer rather than as struct annotations,
// which is idiomatic for Go's database/sql approach.
//
// MIGRATION_NOTE: The Java entity used the following column names:
//   id       -> "User_id"      (auto-generated primary key)
//   name     -> "User_name"    (@NotBlank)
//   email    -> "User_Email"   (unique)
//   password -> "User_Password"
//   role     -> "User_Role"
//   about    -> "User_About"   (length 500)
// The repository layer targets PostgreSQL and should use lower_snake_case
// columns (user_id, user_name, ...) with a GENERATED ALWAYS AS IDENTITY / SERIAL
// primary key and `RETURNING id` on INSERT.
//
// MIGRATION_NOTE: If V-SCHEMA reveals any of these columns are nullable, the
// corresponding field(s) must be changed to *string. As migrated they are all
// non-pointer strings, matching the source which treated them as plain values.
type User struct {
	// ID is the primary key. Zero indicates an unsaved user; the database
	// assigns the value on insert (Postgres IDENTITY/SERIAL).
	ID int `json:"id"`

	// Name is the user's display name. The source marked this @NotBlank with
	// the message "please Add the department Name"; validation is enforced by
	// Validate rather than a struct annotation.
	Name string `json:"name"`

	// Email is the user's unique email address (unique constraint in the DB).
	Email string `json:"email"`

	// Password is the user's (hashed) password.
	Password string `json:"password"`

	// Role is the user's role within the application.
	Role string `json:"role"`

	// About is a free-form profile description (max 500 chars in the source).
	About string `json:"about"`
}

// NewUser constructs a User from the provided fields. It mirrors the Java
// @AllArgsConstructor. All five non-id fields are copied verbatim, including
// empty strings; the caller is responsible for any validation via Validate.
func NewUser(name, email, password, role, about string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate checks the User's fields against the source's Bean Validation
// constraints. It returns an *ErrorMessage describing the first violation, or
// nil if the user is valid.
//
// MIGRATION_NOTE: The Java @NotBlank on name treated a blank (whitespace-only)
// value as invalid. Go's strings.TrimSpace is used to replicate that behaviour.
func (u *User) Validate() *ErrorMessage {
	if isBlank(u.Name) {
		return NewErrorMessage(400, "please Add the department Name")
	}
	return nil
}

// isBlank reports whether s is empty or consists solely of whitespace,
// matching the semantics of Java's @NotBlank.
func isBlank(s string) bool {
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			continue
		default:
			return false
		}
	}
	return true
}
