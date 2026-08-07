package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartErr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the persistence operations available for User
// entities. It mirrors the subset of Spring Data's JpaRepository that the
// application actually uses, plus the derived finder findByName.
//
// MIGRATION_NOTE: Defined as an interface (rather than a concrete type only)
// so that callers depend on an abstraction and the implementation can be
// swapped in tests. The Spring @Repository stereotype (component scanning and
// exception translation) has no direct Go equivalent; wiring is explicit via
// NewUserRepository.
type UserRepository interface {
	// FindByID returns the user with the given id. It returns ErrUserNotFound
	// (wrapped) when no row matches.
	FindByID(ctx context.Context, id int) (model.User, error)
	// FindByName returns the user with the given name. The bool result reports
	// whether a match was found; a false result with a nil error means no such
	// user exists.
	FindByName(ctx context.Context, name string) (model.User, bool, error)
	// FindAll returns all users.
	FindAll(ctx context.Context) ([]model.User, error)
	// Save inserts a new user (when ID is zero) or updates an existing one, and
	// returns the persisted user with its generated ID populated.
	Save(ctx context.Context, user model.User) (model.User, error)
	// DeleteByID removes the user with the given id. It returns ErrUserNotFound
	// (wrapped) when no row matches.
	DeleteByID(ctx context.Context, id int) error
}

// userRepository is the database/sql-backed implementation of UserRepository.
type userRepository struct {
	db *sql.DB
}

// NewUserRepository constructs a UserRepository backed by the given database
// handle. It returns an error when db is nil so misconfiguration is caught at
// wiring time rather than on first use.
func NewUserRepository(db *sql.DB) (UserRepository, error) {
	if db == nil {
		return nil, errors.New("repository: db handle must not be nil")
	}
	return &userRepository{db: db}, nil
}

// FindByID returns the user with the given id.
//
// MIGRATION_NOTE: Per the migration debate notes, the by-id lookup path has
// confirmed 404 behavior, so a missing row maps to ErrUserNotFound.
func (r *userRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	const query = `SELECT id, name, email FROM users WHERE id = $1`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name, &u.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return model.User{}, smartErr.WrapUserNotFound(fmt.Sprintf("user id %d", id), fmt.Errorf("user id %d not found", id))
	case err != nil:
		return model.User{}, fmt.Errorf("repository: find user by id %d: %w", id, err)
	}
	return u, nil
}

// FindByName returns the user with the given name.
//
// MIGRATION_NOTE: This is the direct translation of the derived query
// findByName -> WHERE name = ?. Per the migration debate notes, the source
// behavior for a missing name is unconfirmed, so no not-found policy is baked
// in here: a missing row yields (User{}, false, nil), leaving the decision
// (404, empty result, etc.) to the service layer.
func (r *userRepository) FindByName(ctx context.Context, name string) (model.User, bool, error) {
	const query = `SELECT id, name, email FROM users WHERE name = $1`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, name).Scan(&u.ID, &u.Name, &u.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return model.User{}, false, nil
	case err != nil:
		return model.User{}, false, fmt.Errorf("repository: find user by name %q: %w", name, err)
	}
	return u, true, nil
}

// FindAll returns all users ordered by id.
func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `SELECT id, name, email FROM users ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("repository: scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate user rows: %w", err)
	}
	return users, nil
}

// Save inserts a new user when user.ID is zero, or updates the existing row
// otherwise, returning the persisted user with a populated ID.
//
// MIGRATION_NOTE: JpaRepository.save transparently performs insert-or-update.
// This is replicated explicitly. For inserts we use PostgreSQL's RETURNING id
// (there is no LastInsertId equivalent for the Postgres drivers) to obtain the
// generated identifier.
func (r *userRepository) Save(ctx context.Context, user model.User) (model.User, error) {
	if user.ID == 0 {
		const insert = `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
		if err := r.db.QueryRowContext(ctx, insert, user.Name, user.Email).Scan(&user.ID); err != nil {
			return model.User{}, fmt.Errorf("repository: insert user: %w", err)
		}
		return user, nil
	}

	const update = `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, update, user.Name, user.Email, user.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("repository: update user id %d: %w", user.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return model.User{}, fmt.Errorf("repository: rows affected for user id %d: %w", user.ID, err)
	}
	if affected == 0 {
		return model.User{}, smartErr.WrapUserNotFound(fmt.Sprintf("user id %d", user.ID), fmt.Errorf("user id %d not found", user.ID))
	}
	return user, nil
}

// DeleteByID removes the user with the given id.
func (r *userRepository) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user id %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected for delete user id %d: %w", id, err)
	}
	if affected == 0 {
		return smartErr.WrapUserNotFound(fmt.Sprintf("user id %d", id), fmt.Errorf("user id %d not found", id))
	}
	return nil
}