// Package user contains the User domain model, its validation rules and the
// MySQL schema that mirrors the Hibernate-generated `user` table.
package user

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// User is a registered account of the Smart Contact Manager.
//
// String fields are pointers so that SQL NULL / JSON null round-trip exactly
// like the nullable java.lang.String fields of the original JPA entity.
// Password is intentionally serialised (the source has no @JsonIgnore);
// it is expected to already be BCrypt-hashed by the service layer.
type User struct {
	ID       int     `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
	About    *string `json:"about"`
}

// NewUser returns an empty User (equivalent of the Lombok no-args constructor).
func NewUser() *User {
	return &User{}
}

// NewUserWith returns a User with every field set, in declaration order
// (equivalent of the Lombok all-args constructor).
func NewUserWith(id int, name, email, password, role, about *string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// NameRequiredMessage is the verbatim @NotBlank message from the source entity.
const NameRequiredMessage = "please Add the department Name"

// ValidationError describes a single violated field constraint.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// Validate reproduces the Bean Validation constraints of the source entity:
// name must be non-nil and contain at least one non-whitespace character.
// Callers should invoke it on create (where @Valid was used), not on update.
func (u *User) Validate() error {
	if u.Name == nil || strings.TrimSpace(*u.Name) == "" {
		return &ValidationError{Field: "name", Message: NameRequiredMessage}
	}
	return nil
}

// Equal reports value equality over all fields (Lombok @Data equals).
func (u *User) Equal(o *User) bool {
	if u == nil || o == nil {
		return u == o
	}
	return u.ID == o.ID &&
		strPtrEqual(u.Name, o.Name) &&
		strPtrEqual(u.Email, o.Email) &&
		strPtrEqual(u.Password, o.Password) &&
		strPtrEqual(u.Role, o.Role) &&
		strPtrEqual(u.About, o.About)
}

// String renders the user in Lombok's toString format, e.g.
// "User(id=1, name=Bob, email=null, ...)".
func (u *User) String() string {
	if u == nil {
		return "null"
	}
	var b strings.Builder
	b.WriteString("User(id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString(", name=")
	b.WriteString(strPtrString(u.Name))
	b.WriteString(", email=")
	b.WriteString(strPtrString(u.Email))
	b.WriteString(", password=")
	b.WriteString(strPtrString(u.Password))
	b.WriteString(", role=")
	b.WriteString(strPtrString(u.Role))
	b.WriteString(", about=")
	b.WriteString(strPtrString(u.About))
	b.WriteString(")")
	return b.String()
}

func strPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func strPtrString(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}

// Builder is a fluent builder for User (equivalent of Lombok @Builder).
type Builder struct {
	u User
}

// NewBuilder starts building a User.
func NewBuilder() *Builder {
	return &Builder{}
}

// ID sets the id.
func (b *Builder) ID(id int) *Builder { b.u.ID = id; return b }

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

// Build returns a copy of the User assembled so far.
func (b *Builder) Build() *User {
	u := b.u
	return &u
}

// Table and column names, matching Hibernate's SpringPhysicalNamingStrategy
// output for the source @Table/@Column annotations (lower-cased).
const (
	TableName      = "user"
	ColumnID       = "user_id"
	ColumnName     = "user_name"
	ColumnEmail    = "user_email"
	ColumnPassword = "user_password"
	ColumnRole     = "user_role"
	ColumnAbout    = "user_about"
)

// schemaStatements mirror what Hibernate 5 (MySQL8Dialect, ddl-auto=update)
// creates for this entity. GenerationType.AUTO resolves to a
// `hibernate_sequence` table, so user_id has no AUTO_INCREMENT: ids are
// allocated from hibernate_sequence by the repository, keeping existing
// Hibernate-created databases byte-compatible.
//
// MIGRATION_NOTE: the target-dialect guidance prefers AUTO_INCREMENT, but the
// agreed project plan requires hibernate_sequence id allocation for
// compatibility with databases already created by Hibernate.
var schemaStatements = []string{
	"CREATE TABLE IF NOT EXISTS `user` (" +
		"`user_id` INT NOT NULL, " +
		"`user_about` VARCHAR(500), " +
		"`user_email` VARCHAR(255), " +
		"`user_name` VARCHAR(255), " +
		"`user_password` VARCHAR(255), " +
		"`user_role` VARCHAR(255), " +
		"PRIMARY KEY (`user_id`), " +
		"UNIQUE KEY `UK_user_email` (`user_email`)" +
		") ENGINE=InnoDB",
	"CREATE TABLE IF NOT EXISTS `hibernate_sequence` (`next_val` BIGINT) ENGINE=InnoDB",
	"INSERT INTO `hibernate_sequence` (`next_val`) " +
		"SELECT 1 FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `hibernate_sequence`)",
}

// EnsureSchema creates the user table and the hibernate_sequence id table if
// they do not exist. It must be called once at application start-up
// (replacement for spring.jpa.hibernate.ddl-auto=update).
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	for _, stmt := range schemaStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("user: ensure schema: %w", err)
		}
	}
	return nil
}
