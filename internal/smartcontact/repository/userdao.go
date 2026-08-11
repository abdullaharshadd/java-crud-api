// Package repository contains the data-access layer migrated from the
// Spring Data JPA repositories under com.smartContact.repository.
//
// MIGRATION_NOTE: The Java UserDao was a Spring Data JpaRepository<User,
// Integer> interface with no implementation body — Spring synthesized the
// CRUD methods and derived a query from the findByName method name at
// runtime. Go has no such runtime proxy generation, so this file provides
// a concrete UserDao backed by database/sql with explicit SQL for every
// operation the source depended on (save/create, findById, findByName,
// delete). Table and column names follow the migrated model.User schema
// ("users" table, lower_snake_case columns) and PostgreSQL positional
// placeholders ($1, $2, ...) per the target dialect.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// ErrUserNotFound is returned by mutating operations (e.g. Delete) when the
// targeted row does not exist.
//
// MIGRATION_NOTE: The migration notes require Delete to return a generic
// (non-nil) error when no row matches, whereas GetUserNameByName must
// return (nil, nil) on a name miss. This sentinel satisfies the former;
// callers that want a 404-flavored error can wrap it with apperr.
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the persistence operations the service layer needs
// for User entities. It replaces the Spring Data JpaRepository<User,
// Integer> interface; only the methods the application actually used are
// exposed rather than the full JpaRepository surface.
type UserRepository interface {
	// Save inserts a new user (when ID is zero) or updates an existing one,
	// mirroring JpaRepository.save. It returns the persisted user with its
	// generated ID populated on insert.
	Save(ctx context.Context, u *model.User) (*model.User, error)
	// FindByID looks up a user by primary key, mirroring
	// JpaRepository.findById. It returns (nil, false, nil) when no row exists.
	FindByID(ctx context.Context, id int) (*model.User, bool, error)
	// FindByName looks up a user by name, mirroring the derived query
	// findByName. It returns (nil, nil) when no row matches.
	FindByName(ctx context.Context, name string) (*model.User, error)
	// Delete removes the user with the given ID, mirroring
	// JpaRepository.deleteById. It returns ErrUserNotFound when no row exists.
	Delete(ctx context.Context, id int) error
}

// userDao is the database/sql-backed implementation of UserRepository.
//
// MIGRATION_NOTE: Named userDao to preserve traceability to the source
// UserDao interface; the exported abstraction is UserRepository.
type userDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserRepository backed by the given *sql.DB.
func NewUserDao(db *sql.DB) UserRepository {
	return &userDao{db: db}
}

// Save inserts or updates a user. When u.ID is zero the row is inserted and
// the DB-generated identity is scanned back via RETURNING id; otherwise the
// existing row is updated in place.
func (d *userDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("repository: cannot save nil user")
	}

	if u.ID == 0 {
		const insertSQL = `INSERT INTO users (name, email, password, role, about)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`
		if err := d.db.QueryRowContext(ctx, insertSQL,
			u.Name, u.Email, u.Password, u.Role, u.About,
		).Scan(&u.ID); err != nil {
			return nil, fmt.Errorf("repository: insert user: %w", err)
		}
		return u, nil
	}

	const updateSQL = `UPDATE users
		SET name = $1, email = $2, password = $3, role = $4, about = $5
		WHERE id = $6`
	res, err := d.db.ExecContext(ctx, updateSQL,
		u.Name, u.Email, u.Password, u.Role, u.About, u.ID)
	if err != nil {
		return nil, fmt.Errorf("repository: update user %d: %w", u.ID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository: rows affected for user %d: %w", u.ID, err)
	}
	if n == 0 {
		return nil, apperr.NewUserNotFoundErrorf("no user with id %d to update", u.ID)
	}
	return u, nil
}

// FindByID looks up a user by primary key. The bool result is false (with a
// nil error) when no row matches, mirroring JpaRepository.findById returning
// an empty Optional.
func (d *userDao) FindByID(ctx context.Context, id int) (*model.User, bool, error) {
	const querySQL = `SELECT id, name, email, password, role, about
		FROM users
		WHERE id = $1`
	u, err := d.scanUser(d.db.QueryRowContext(ctx, querySQL, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("repository: find user by id %d: %w", id, err)
	}
	return u, true, nil
}

// FindByName implements the derived findByName query.
//
// MIGRATION_NOTE: Per the migration notes, a name miss returns (nil, nil)
// rather than an error, matching the source's derived finder which returned
// null when no matching row existed.
func (d *userDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const querySQL = `SELECT id, name, email, password, role, about
		FROM users
		WHERE name = $1`
	u, err := d.scanUser(d.db.QueryRowContext(ctx, querySQL, name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: find user by name %q: %w", name, err)
	}
	return u, nil
}

// Delete removes the user with the given ID.
//
// MIGRATION_NOTE: Per the migration notes, deleting a missing row returns a
// generic (non-nil) error (ErrUserNotFound) rather than silently
// succeeding, unlike JpaRepository.deleteById which throws
// EmptyResultDataAccessException.
func (d *userDao) Delete(ctx context.Context, id int) error {
	const deleteSQL = `DELETE FROM users WHERE id = $1`
	res, err := d.db.ExecContext(ctx, deleteSQL, id)
	if err != nil {
		return fmt.Errorf("repository: delete user %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected deleting user %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("repository: delete user %d: %w", id, ErrUserNotFound)
	}
	return nil
}

// rowScanner abstracts *sql.Row / *sql.Rows for the shared scan helper.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser maps a single result row onto a model.User.
func (d *userDao) scanUser(row rowScanner) (*model.User, error) {
	var u model.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About); err != nil {
		return nil, err
	}
	return &u, nil
}
