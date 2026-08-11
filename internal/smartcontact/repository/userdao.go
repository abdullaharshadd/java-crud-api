// Package repository provides data-access implementations for the smartcontact
// application. It replaces Spring Data JPA repositories with explicit
// database/sql-backed methods against PostgreSQL.
//
// MIGRATION_NOTE: The Java source was a Spring Data JPA interface
// (UserDao extends JpaRepository<User, Integer>) with a single derived query
// method findByName. Spring auto-generated the CRUD implementation and the
// findByName proxy at runtime. Go has no equivalent runtime proxy mechanism,
// so we implement a concrete UserDao struct that owns a *sql.DB and issues raw
// SQL. The inherited JpaRepository CRUD surface (save, findById, findAll,
// deleteById, existsById, count) is realized as explicit methods so callers
// that relied on those operations continue to work.
//
// Per the migration debate notes, a sentinel ErrNoRowsDeleted replicates
// Spring's EmptyResultDataAccessException behavior when deleting a
// non-existent id.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/smartContact/internal/smartcontact/model"
)

// ErrNoRowsDeleted is returned by DeleteByID when no row matches the supplied
// id. It mirrors the Spring Data behavior where deleteById on a missing id
// raised EmptyResultDataAccessException.
//
// MIGRATION_NOTE: The exact HTTP mapping (Spring produced a 500) is a
// presentation concern handled by the HTTP middleware, not by the repository.
// Callers can detect this condition with errors.Is(err, ErrNoRowsDeleted).
var ErrNoRowsDeleted = errors.New("repository: no rows deleted")

// ErrUserNotFound is returned by read operations that expect exactly one row
// but find none. It lets callers distinguish "not found" from other failures
// via errors.Is without importing database/sql.
var ErrUserNotFound = errors.New("repository: user not found")

// createUsersTableDDL creates the users table if it does not already exist.
//
// MIGRATION_NOTE: The Java source relied on Hibernate's ddl-auto to derive the
// schema from the @Entity mapping. Since this project runs no external
// migration tool, we create the real schema at startup here. Column names and
// types are derived from the migrated model.User struct (lower_snake_case for
// PostgreSQL). The primary key uses GENERATED ALWAYS AS IDENTITY so ID is
// assigned by the database and returned via RETURNING id on INSERT.
const createUsersTableDDL = `
CREATE TABLE IF NOT EXISTS users (
	id    INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name  TEXT NOT NULL
)`

// UserDao provides CRUD and query access to User records backed by a
// PostgreSQL database. Construct one with NewUserDao.
type UserDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserDao backed by the given *sql.DB and ensures the
// underlying schema exists. It returns an error if the schema cannot be
// created.
//
// MIGRATION_NOTE: Schema creation is performed at construction time to replace
// Hibernate's ddl-auto=create/update behavior. Pass a context that bounds the
// startup DDL execution.
func NewUserDao(ctx context.Context, db *sql.DB) (*UserDao, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	if _, err := db.ExecContext(ctx, createUsersTableDDL); err != nil {
		return nil, fmt.Errorf("creating users table: %w", err)
	}
	return &UserDao{db: db}, nil
}

// FindByName returns the user with the given name.
//
// MIGRATION_NOTE: This is the direct equivalent of the Spring Data derived
// query method `User findByName(String name)`. Spring returned null when no
// match was found; idiomatic Go returns (nil, ErrUserNotFound) instead so the
// caller must handle the missing case explicitly.
func (d *UserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE name = $1`

	var u model.User
	err := d.db.QueryRowContext(ctx, query, name).Scan(&u.ID, &u.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("finding user by name %q: %w", name, err)
	}
	return &u, nil
}

// FindByID returns the user with the given id, replacing JpaRepository.findById.
// It returns ErrUserNotFound when no row matches.
func (d *UserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE id = $1`

	var u model.User
	err := d.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("finding user by id %d: %w", id, err)
	}
	return &u, nil
}

// FindAll returns every user, replacing JpaRepository.findAll.
func (d *UserDao) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `SELECT id, name FROM users ORDER BY id`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}
	return users, nil
}

// Save inserts a new user or updates an existing one, replacing
// JpaRepository.save. When u.ID is zero the row is inserted and the
// database-generated id is populated on the returned user via RETURNING id;
// otherwise the existing row is updated.
func (d *UserDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("repository: user must not be nil")
	}

	if u.ID == 0 {
		const insert = `INSERT INTO users (name) VALUES ($1) RETURNING id`
		if err := d.db.QueryRowContext(ctx, insert, u.Name).Scan(&u.ID); err != nil {
			return nil, fmt.Errorf("inserting user: %w", err)
		}
		return u, nil
	}

	const update = `UPDATE users SET name = $1 WHERE id = $2`
	res, err := d.db.ExecContext(ctx, update, u.Name, u.ID)
	if err != nil {
		return nil, fmt.Errorf("updating user %d: %w", u.ID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("checking rows affected for user %d: %w", u.ID, err)
	}
	if n == 0 {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// DeleteByID removes the user with the given id, replacing
// JpaRepository.deleteById. It returns ErrNoRowsDeleted when no row matches,
// mirroring Spring's EmptyResultDataAccessException on a missing id.
func (d *UserDao) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`

	res, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting user %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for user %d: %w", id, err)
	}
	if n == 0 {
		return ErrNoRowsDeleted
	}
	return nil
}

// ExistsByID reports whether a user with the given id exists, replacing
// JpaRepository.existsById.
func (d *UserDao) ExistsByID(ctx context.Context, id int) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	if err := d.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking existence of user %d: %w", id, err)
	}
	return exists, nil
}

// Count returns the total number of users, replacing JpaRepository.count.
func (d *UserDao) Count(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM users`

	var n int64
	if err := d.db.QueryRowContext(ctx, query).Scan(&n); err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return n, nil
}
