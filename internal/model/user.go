// Package model contains the persistent domain types of the application.
package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// UserNameBlankMessage is the @NotBlank message attached to User.Name in the
// source entity. It is used for logs only.
const UserNameBlankMessage = "please Add the department Name"

// ErrUserNameBlank is returned by User.Validate when the name is nil or
// whitespace-only (Bean Validation @NotBlank semantics).
var ErrUserNameBlank = errors.New(UserNameBlankMessage)

// User table and column names. Hibernate's SpringPhysicalNamingStrategy
// lower-cases the explicit @Table/@Column names of the source entity.
const (
	UserTable         = "user"
	UserColumnID      = "user_id"
	UserColumnName    = "user_name"
	UserColumnEmail   = "user_email"
	UserColumnPass    = "user_password"
	UserColumnRole    = "user_role"
	UserColumnAbout   = "user_about"
	UserAboutMaxLen   = 500
	HibernateSequence = "hibernate_sequence"
)

// userSchemaDDL is the idempotent DDL equivalent to what Hibernate's
// ddl-auto generates for the User entity with GenerationType.AUTO on MySQL
// (a hibernate_sequence table holding the next id, plus the user table).
//
// MIGRATION_NOTE: ids are allocated from hibernate_sequence by the store
// layer to mirror Hibernate; AUTO_INCREMENT is kept on the PK as a harmless
// fallback for inserts that omit the id.
var userSchemaDDL = []string{
	"CREATE TABLE IF NOT EXISTS `hibernate_sequence` (`next_val` BIGINT) ENGINE=InnoDB",
	"INSERT INTO `hibernate_sequence` (`next_val`) SELECT 1 FROM DUAL " +
		"WHERE NOT EXISTS (SELECT 1 FROM `hibernate_sequence`)",
	"CREATE TABLE IF NOT EXISTS `user` (" +
		"`user_id` INT NOT NULL AUTO_INCREMENT, " +
		"`user_about` VARCHAR(500) NULL, " +
		"`user_email` VARCHAR(255) NULL, " +
		"`user_name` VARCHAR(255) NULL, " +
		"`user_password` VARCHAR(255) NULL, " +
		"`user_role` VARCHAR(255) NULL, " +
		"PRIMARY KEY (`user_id`), " +
		"UNIQUE KEY `uk_user_email` (`user_email`)" +
		") ENGINE=InnoDB",
}

// UserSchemaDDL returns a copy of the statements that create the user schema.
func UserSchemaDDL() []string {
	out := make([]string, len(userSchemaDDL))
	copy(out, userSchemaDDL)
	return out
}

// Execer is the subset of *sql.DB / *sql.Tx needed to run DDL.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// EnsureUserSchema creates the hibernate_sequence and user tables if they do
// not exist yet. It is safe to call on every startup.
func EnsureUserSchema(ctx context.Context, db Execer) error {
	if db == nil {
		return errors.New("model: ensure user schema: nil database handle")
	}
	for _, stmt := range userSchemaDDL {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("model: ensure user schema: %w", err)
		}
	}
	return nil
}

// User is the persisted user entity (source: JPA entity mapped to table USER).
//
// String fields are pointers so that SQL NULL / JSON null stay distinct from
// the empty string, matching Java's nullable String. The password is exposed
// in JSON exactly as in the source.
type User struct {
	ID       int32   `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
	About    *string `json:"about"`
}

// NewEmptyUser returns a User with every field at its zero value
// (equivalent of the no-args constructor).
func NewEmptyUser() *User {
	return &User{}
}

// NewUser returns a User with every field set, in declaration order
// (equivalent of the all-args constructor).
func NewUser(id int32, name, email, password, role, about *string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate enforces the Bean Validation constraints of the source entity:
// name must be non-nil and contain at least one non-whitespace character.
func (u *User) Validate() error {
	if u == nil || u.Name == nil || strings.TrimSpace(*u.Name) == "" {
		return ErrUserNameBlank
	}
	return nil
}

// GetID returns the user id.
func (u *User) GetID() int32 { return u.ID }

// SetID sets the user id.
func (u *User) SetID(id int32) { u.ID = id }

// GetName returns the user name (nil when unset).
func (u *User) GetName() *string { return u.Name }

// SetName sets the user name.
func (u *User) SetName(name *string) { u.Name = name }

// GetEmail returns the user email (nil when unset).
func (u *User) GetEmail() *string { return u.Email }

// SetEmail sets the user email.
func (u *User) SetEmail(email *string) { u.Email = email }

// GetPassword returns the user password (nil when unset).
func (u *User) GetPassword() *string { return u.Password }

// SetPassword sets the user password.
func (u *User) SetPassword(password *string) { u.Password = password }

// GetRole returns the user role (nil when unset).
func (u *User) GetRole() *string { return u.Role }

// SetRole sets the user role.
func (u *User) SetRole(role *string) { u.Role = role }

// GetAbout returns the user "about" text (nil when unset).
func (u *User) GetAbout() *string { return u.About }

// SetAbout sets the user "about" text.
func (u *User) SetAbout(about *string) { u.About = about }

// Equal reports whether u and other have equal values for all six fields
// (Lombok @Data equals semantics: nil equals nil, otherwise string contents).
func (u *User) Equal(other *User) bool {
	if u == nil || other == nil {
		return u == other
	}
	return u.ID == other.ID &&
		equalStrPtr(u.Name, other.Name) &&
		equalStrPtr(u.Email, other.Email) &&
		equalStrPtr(u.Password, other.Password) &&
		equalStrPtr(u.Role, other.Role) &&
		equalStrPtr(u.About, other.About)
}

// String renders every field in Lombok's toString format, e.g.
// "User(id=1, name=a, email=null, password=..., role=..., about=...)".
func (u *User) String() string {
	if u == nil {
		return "null"
	}
	return fmt.Sprintf("User(id=%d, name=%s, email=%s, password=%s, role=%s, about=%s)",
		u.ID, strOrNull(u.Name), strOrNull(u.Email), strOrNull(u.Password),
		strOrNull(u.Role), strOrNull(u.About))
}

// UserBuilder is a fluent builder for User (equivalent of Lombok @Builder).
type UserBuilder struct {
	u User
}

// NewUserBuilder starts building a User.
func NewUserBuilder() *UserBuilder { return &UserBuilder{} }

// ID sets the id.
func (b *UserBuilder) ID(id int32) *UserBuilder { b.u.ID = id; return b }

// Name sets the name.
func (b *UserBuilder) Name(name string) *UserBuilder { b.u.Name = &name; return b }

// Email sets the email.
func (b *UserBuilder) Email(email string) *UserBuilder { b.u.Email = &email; return b }

// Password sets the password.
func (b *UserBuilder) Password(password string) *UserBuilder { b.u.Password = &password; return b }

// Role sets the role.
func (b *UserBuilder) Role(role string) *UserBuilder { b.u.Role = &role; return b }

// About sets the about text.
func (b *UserBuilder) About(about string) *UserBuilder { b.u.About = &about; return b }

// Build returns a new User with the values accumulated so far.
func (b *UserBuilder) Build() *User {
	u := b.u
	return &u
}

func equalStrPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func strOrNull(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}
