package model

import (
	"strconv"
	"strings"
)

// User maps to the "USER" table and holds user identity, credentials,
// role, and profile fields.
//
// MIGRATION_NOTE: The source stored and exposed the password in plaintext
// (no @JsonIgnore, no hashing). This is preserved faithfully here —
// Password is serialized on every GET. This is a security issue and must
// be addressed in a hardening pass (hash at write time, omit from reads).
//
// MIGRATION_NOTE: ID is a non-pointer int so the zero value is 0 (matching
// Java's int default). All other fields are *string so an omitted/null JSON
// value maps to SQL NULL rather than an empty string.
//
// MIGRATION_NOTE: Schema creation (Hibernate ddl-auto=update) is NOT handled
// here — in idiomatic Go the DDL lives in the DB/repository layer (db.go /
// repository). The source table columns are: User_id (PK, AUTO_INCREMENT),
// User_name, User_Email (UNIQUE), User_Password, User_Role, User_About
// (VARCHAR(500)).
type User struct {
	ID       int     `json:"id" db:"User_id"`
	Name     *string `json:"name" db:"User_name" validate:"required"`
	Email    *string `json:"email" db:"User_Email"`
	Password *string `json:"password" db:"User_Password"`
	Role     *string `json:"role" db:"User_Role"`
	About    *string `json:"about" db:"User_About"`
}

// NewUser constructs a User with all fields populated in declaration order,
// mirroring the source's all-args constructor.
func NewUser(id int, name, email, password, role, about *string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// UserBuilder constructs a User incrementally by named fields, mirroring the
// source's Lombok @Builder.
type UserBuilder struct {
	user User
}

// NewUserBuilder returns an empty UserBuilder.
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

// ID sets the primary key.
func (b *UserBuilder) ID(id int) *UserBuilder {
	b.user.ID = id
	return b
}

// Name sets the user's name.
func (b *UserBuilder) Name(name string) *UserBuilder {
	b.user.Name = &name
	return b
}

// Email sets the user's email.
func (b *UserBuilder) Email(email string) *UserBuilder {
	b.user.Email = &email
	return b
}

// Password sets the user's password.
func (b *UserBuilder) Password(password string) *UserBuilder {
	b.user.Password = &password
	return b
}

// Role sets the user's role.
func (b *UserBuilder) Role(role string) *UserBuilder {
	b.user.Role = &role
	return b
}

// About sets the user's profile 'about' text.
func (b *UserBuilder) About(about string) *UserBuilder {
	b.user.About = &about
	return b
}

// Build returns the constructed User.
func (b *UserBuilder) Build() *User {
	u := b.user
	return &u
}

// GetID returns the user's primary key id.
func (u *User) GetID() int { return u.ID }

// SetID sets the user's primary key id.
func (u *User) SetID(id int) { u.ID = id }

// GetName returns the user's name, or "" if unset.
func (u *User) GetName() string { return deref(u.Name) }

// SetName sets the user's name.
func (u *User) SetName(name string) { u.Name = &name }

// GetEmail returns the user's email, or "" if unset.
// The email column is unique in the database.
func (u *User) GetEmail() string { return deref(u.Email) }

// SetEmail sets the user's email.
func (u *User) SetEmail(email string) { u.Email = &email }

// GetPassword returns the user's password, or "" if unset.
func (u *User) GetPassword() string { return deref(u.Password) }

// SetPassword sets the user's password.
func (u *User) SetPassword(password string) { u.Password = &password }

// GetRole returns the user's role, or "" if unset.
func (u *User) GetRole() string { return deref(u.Role) }

// SetRole sets the user's role.
func (u *User) SetRole(role string) { u.Role = &role }

// GetAbout returns the user's profile 'about' text, or "" if unset.
// The column is limited to 500 characters in the database.
func (u *User) GetAbout() string { return deref(u.About) }

// SetAbout sets the user's profile 'about' text.
func (u *User) SetAbout(about string) { u.About = &about }

// Equals reports value-based equality across all fields, mirroring Lombok's
// @Data-generated equals.
func (u *User) Equals(other *User) bool {
	if u == nil || other == nil {
		return u == other
	}
	return u.ID == other.ID &&
		ptrEq(u.Name, other.Name) &&
		ptrEq(u.Email, other.Email) &&
		ptrEq(u.Password, other.Password) &&
		ptrEq(u.Role, other.Role) &&
		ptrEq(u.About, other.About)
}

// String returns a representation including all field values, mirroring
// Lombok's @Data-generated toString.
func (u *User) String() string {
	var b strings.Builder
	b.WriteString("User(id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString(", name=")
	b.WriteString(deref(u.Name))
	b.WriteString(", email=")
	b.WriteString(deref(u.Email))
	b.WriteString(", password=")
	b.WriteString(deref(u.Password))
	b.WriteString(", role=")
	b.WriteString(deref(u.Role))
	b.WriteString(", about=")
	b.WriteString(deref(u.About))
	b.WriteString(")")
	return b.String()
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
