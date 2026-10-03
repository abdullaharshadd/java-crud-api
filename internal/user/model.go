// Package user contains the User domain model for the Smart Contact Manager
// application: the account record (id, name, email, password, role, about),
// its validation rules and the relational mapping used to persist it.
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrValidation is returned (wrapped) by Validate when a User violates a
// validation rule. Callers can test for it with errors.Is.
var ErrValidation = errors.New("validation failed")

// NameRequiredMessage is the exact message of the source's
// @NotBlank(message = "please Add the department Name") constraint on name.
const NameRequiredMessage = "please Add the department Name"

// User is an application account.
//
// String fields are pointers so that SQL NULL / JSON null stays distinct from
// the empty string, mirroring java.lang.String semantics in the source entity.
// JSON shape is {"id","name","email","password","role","about"}; a null or
// absent id decodes as 0 and unknown fields are ignored.
//
// MIGRATION_NOTE: encoding/json matches object keys case-insensitively
// (e.g. "NAME" populates Name), whereas Jackson is case-sensitive. This is an
// accepted divergence.
//
// MIGRATION_NOTE: Password is stored and serialized in plaintext exactly as in
// the source. Flagged for follow-up (hashing, omitting from responses).
type User struct {
	ID       int32   `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
	About    *string `json:"about"`
}

// NewUser creates a User with every field supplied positionally, in the order
// id, name, email, password, role, about (Lombok @AllArgsConstructor).
// A nil string pointer represents a null value.
func NewUser(id int32, name, email, password, role, about *string) User {
	return User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// StringPtr returns a pointer to s; a convenience for building Users.
func StringPtr(s string) *string { return &s }

// Validate applies the bean-validation rules of the source entity:
// Name must be non-nil and non-blank (@NotBlank). Blankness follows Java's
// String.trim semantics: leading/trailing runes <= U+0020 are ignored.
// Other fields are not validated (email has no @Email constraint).
func (u User) Validate() error {
	if u.Name == nil || isJavaBlank(*u.Name) {
		return fmt.Errorf("%w: name: %s", ErrValidation, NameRequiredMessage)
	}
	return nil
}

func isJavaBlank(s string) bool {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' }) == ""
}

// Equal reports whether u and other have equal values in all six fields
// (Lombok @Data equals). Nil string pointers are equal only to nil.
func (u User) Equal(other User) bool {
	return u.ID == other.ID &&
		strPtrEqual(u.Name, other.Name) &&
		strPtrEqual(u.Email, other.Email) &&
		strPtrEqual(u.Password, other.Password) &&
		strPtrEqual(u.Role, other.Role) &&
		strPtrEqual(u.About, other.About)
}

func strPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// String returns a representation including every field, in Lombok's
// format: User(id=1, name=..., email=..., password=..., role=..., about=...).
// Nil fields print as "null".
func (u User) String() string {
	var b strings.Builder
	b.WriteString("User(id=")
	b.WriteString(strconv.FormatInt(int64(u.ID), 10))
	b.WriteString(", name=")
	b.WriteString(strOrNull(u.Name))
	b.WriteString(", email=")
	b.WriteString(strOrNull(u.Email))
	b.WriteString(", password=")
	b.WriteString(strOrNull(u.Password))
	b.WriteString(", role=")
	b.WriteString(strOrNull(u.Role))
	b.WriteString(", about=")
	b.WriteString(strOrNull(u.About))
	b.WriteString(")")
	return b.String()
}

func strOrNull(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}

// Builder is a fluent builder for User (Lombok @Builder). Fields that are not
// set keep their zero/null value.
type Builder struct {
	u User
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder { return &Builder{} }

// ID sets the id.
func (b *Builder) ID(id int32) *Builder { b.u.ID = id; return b }

// Name sets the name.
func (b *Builder) Name(v string) *Builder { b.u.Name = &v; return b }

// Email sets the email.
func (b *Builder) Email(v string) *Builder { b.u.Email = &v; return b }

// Password sets the password.
func (b *Builder) Password(v string) *Builder { b.u.Password = &v; return b }

// Role sets the role.
func (b *Builder) Role(v string) *Builder { b.u.Role = &v; return b }

// About sets the about text.
func (b *Builder) About(v string) *Builder { b.u.About = &v; return b }

// Build returns the constructed User. The builder may be reused; each setter
// allocates a fresh string, so later calls do not affect already built Users.
func (b *Builder) Build() User { return b.u }

// Persistence mapping (source: @Entity @Table(name="USER") with Spring's
// physical naming strategy, which lower-cases identifiers).
const (
	// TableName is the table storing users.
	TableName = "user"
	// ColID is the primary-key column.
	ColID = "user_id"
	// ColName is the name column.
	ColName = "user_name"
	// ColEmail is the unique email column.
	ColEmail = "user_email"
	// ColPassword is the password column.
	ColPassword = "user_password"
	// ColRole is the role column.
	ColRole = "user_role"
	// ColAbout is the about column (length 500).
	ColAbout = "user_about"
)

// SchemaStatements is the DDL equivalent of what Hibernate
// (ddl-auto=update, MySQL8Dialect) generates for this entity.
//
// MIGRATION_NOTE: @GeneratedValue(strategy = AUTO) on Hibernate 5 + MySQL
// uses a hibernate_sequence table rather than AUTO_INCREMENT, so user_id is a
// plain INT primary key and ids are allocated from hibernate_sequence by the
// repository layer. Switching to AUTO_INCREMENT would change id allocation
// behavior and is left for follow-up.
var SchemaStatements = []string{
	"CREATE TABLE IF NOT EXISTS `user` (" +
		"`user_id` INT NOT NULL, " +
		"`user_about` VARCHAR(500) NULL, " +
		"`user_email` VARCHAR(255) NULL, " +
		"`user_name` VARCHAR(255) NULL, " +
		"`user_password` VARCHAR(255) NULL, " +
		"`user_role` VARCHAR(255) NULL, " +
		"PRIMARY KEY (`user_id`), " +
		"UNIQUE KEY `uk_user_email` (`user_email`)" +
		") ENGINE=InnoDB",
	"CREATE TABLE IF NOT EXISTS `hibernate_sequence` (`next_val` BIGINT) ENGINE=InnoDB",
	"INSERT INTO `hibernate_sequence` (`next_val`) " +
		"SELECT 1 FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `hibernate_sequence`)",
}

// Execer is the subset of *sql.DB / *sql.Tx needed to apply the schema.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// EnsureSchema creates the user table and the hibernate_sequence table (seeded
// with 1) if they do not exist. Each statement runs separately because the
// MySQL driver does not allow multiple statements per Exec by default.
func EnsureSchema(ctx context.Context, db Execer) error {
	for i, stmt := range SchemaStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure user schema (statement %d): %w", i+1, err)
		}
	}
	return nil
}
