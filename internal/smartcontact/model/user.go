// Package model defines the domain entities for the smartContact application.
package model

// User represents an application user persisted in the USER database table.
//
// MIGRATION_NOTE: The original Java class relied on JPA annotations (@Entity,
// @Table, @Column) for ORM mapping. In idiomatic Go we keep the persistence
// concern out of the model where possible and use struct tags for whichever
// data-access library is chosen. The `db` tags below match the original column
// names so they can be used with sqlx or similar. If you use GORM instead,
// replace the `db` tags with `gorm:"column:..."` tags and add a TableName
// method. The unique constraint on email (@Column(unique = true)) must be
// enforced in the schema migration / DDL, not in the struct definition.
type User struct {
	// ID is the auto-generated primary key (USER.User_id).
	//
	// MIGRATION_NOTE: JPA used GenerationType.AUTO. In Go the value is generated
	// by the database (e.g. SERIAL / AUTO_INCREMENT) or by the repository layer.
	ID int64 `db:"User_id" json:"id"`

	// Name is the user's display name (USER.User_name).
	Name string `db:"User_name" json:"name"`

	// Email is the user's unique email address (USER.User_Email).
	Email string `db:"User_Email" json:"email"`

	// Password is the user's (hashed) password (USER.User_Password).
	//
	// MIGRATION_NOTE: never serialize the password back to clients. The
	// json:"-" tag ensures it is omitted from any JSON response, and
	// ToResponse below provides an extra guarantee by not copying it.
	Password string `db:"User_Password" json:"-"`

	// Role is the user's authorization role (USER.User_Role).
	Role string `db:"User_Role" json:"role"`

	// About is a free-text description of the user, max 500 characters
	// (USER.User_About).
	About string `db:"User_About" json:"about"`
}

// NewUser constructs a User with the provided field values.
//
// MIGRATION_NOTE: this replaces the Lombok @Builder / @AllArgsConstructor.
// Callers that need only some fields can construct a User literal directly.
func NewUser(id int64, name, email, password, role, about string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// UserRequest is the inbound payload used when creating or updating a user.
// Validation is expressed with go-playground/validator struct tags, replacing
// the Java Bean Validation @NotBlank annotation.
type UserRequest struct {
	// Name must not be blank.
	//
	// MIGRATION_NOTE: the original message was "please Add the department Name",
	// which appears to be a copy-paste artifact referring to a department rather
	// than a user. Preserved logically as a required constraint; human review
	// recommended to confirm the intended validation message.
	Name string `json:"name" validate:"required"`

	// Email must not be blank and must be a valid email address.
	Email string `json:"email" validate:"required,email"`

	// Password must not be blank.
	Password string `json:"password" validate:"required"`

	// Role is optional at the request level.
	Role string `json:"role" validate:"omitempty"`

	// About is optional and limited to 500 characters, mirroring the column length.
	About string `json:"about" validate:"omitempty,max=500"`
}

// ToModel converts a validated UserRequest into a User domain model.
// The ID is left zero so the persistence layer assigns it.
func (r UserRequest) ToModel() *User {
	return &User{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
		Role:     r.Role,
		About:    r.About,
	}
}

// UserResponse is the outbound representation of a user. It deliberately omits
// the password field so handlers never leak credentials to clients.
type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	About string `json:"about"`
}

// ToResponse maps a User model to its safe outbound representation.
//
// MIGRATION_NOTE: per the migration debate, handlers must never return a raw
// model.User. Always call ToResponse before writing to the response.
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
		About: u.About,
	}
}
