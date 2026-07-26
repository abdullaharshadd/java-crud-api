// Package repository provides persistence access for the SmartContact
// domain entities.
//
// MIGRATION_NOTE: The Java source UserDao.java was a Spring Data JPA repository
// interface (`interface UserDao extends JpaRepository<User, Integer>`). Spring
// generated the implementation at runtime, providing full CRUD plus a derived
// query method (`findByName`) whose SQL was inferred from the method name.
//
// Go has no runtime code generation for repositories, so the idiomatic
// translation is:
//
//   - A UserRepository interface documenting the operations the service needs
//     (Save/FindByID/FindAll/DeleteByID/FindByName). This is the subset of
//     JpaRepository actually exercised by the application, expressed
//     explicitly.
//   - A concrete mysqlUserRepo implementation backed by database/sql, with
//     hand-written SQL replacing Hibernate's generated queries.
//   - A constructor NewMySQLUserRepo for dependency injection.
//
// Behavioural fidelity notes (from the migration debate):
//   - DeleteByID checks RowsAffected == 0 and returns ErrEmptyResultDelete.
//     This deliberately replicates Spring Data's
//     EmptyResultDataAccessException, which the original controller did NOT map
//     to 404 and which therefore surfaced as an unhandled HTTP 500 (Change 12).
//     It must NOT be conflated with NotFoundError.
//   - FindByName in Spring Data returned null when no row matched (and could
//     throw on multiple matches). Here a missing row maps to
//     (nil, ErrUserNotFound) so callers can distinguish absence explicitly.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// ErrEmptyResultDelete is returned by DeleteByID when no row matched the given
// id, mirroring Spring Data's EmptyResultDataAccessException.
//
// MIGRATION_NOTE: This is intentionally distinct from NotFoundError. The
// original Java code left this exception unhandled, producing an HTTP 500 on
// delete-of-missing-id. Callers that want a 404 must translate it explicitly.
var ErrEmptyResultDelete = errors.New("empty result data access: no row deleted")

// UserRepository defines the persistence operations for the User entity.
//
// MIGRATION_NOTE: This replaces the inherited JpaRepository<User, Integer>
// surface. Only the operations the application actually uses are declared,
// following the Go convention of small, purpose-driven interfaces.
type UserRepository interface {
	// Save inserts a new user (or updates an existing one when ID is set) and
	// returns the persisted user with any generated fields populated.
	Save(ctx context.Context, u *model.User) (*model.User, error)
	// FindByID returns the user with the given primary key, or ErrUserNotFound
	// if no such user exists.
	FindByID(ctx context.Context, id int) (*model.User, error)
	// FindAll returns all users.
	FindAll(ctx context.Context) ([]model.User, error)
	// DeleteByID deletes the user with the given primary key. It returns
	// ErrEmptyResultDelete if no row was affected.
	DeleteByID(ctx context.Context, id int) error
	// FindByName returns the user with the given name, or ErrUserNotFound if no
	// such user exists. This is the Go equivalent of the derived query method
	// findByName(String).
	FindByName(ctx context.Context, name string) (*model.User, error)
}

// mysqlUserRepo is the database/sql-backed implementation of UserRepository.
type mysqlUserRepo struct {
	db *sql.DB
}

// NewMySQLUserRepo constructs a UserRepository backed by the given *sql.DB.
//
// MIGRATION_NOTE: Where Spring injected a proxy implementing UserDao, here the
// *sql.DB dependency is passed explicitly to the constructor (DI by hand).
func NewMySQLUserRepo(db *sql.DB) UserRepository {
	return &mysqlUserRepo{db: db}
}

// Save inserts the given user and returns it with the generated ID populated.
func (r *mysqlUserRepo) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("repository: cannot save nil user")
	}

	const query = `INSERT INTO user (name) VALUES (?)`
	res, err := r.db.ExecContext(ctx, query, u.Name)
	if err != nil {
		return nil, fmt.Errorf("repository: save user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: save user: read generated id: %w", err)
	}
	u.ID = int(id)

	return u, nil
}

// FindByID returns the user with the given primary key.
func (r *mysqlUserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `SELECT id, name FROM user WHERE id = ?`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, smartErr.NewUserNotFound(fmt.Sprintf("user with id %d not found", id))
	case err != nil:
		return nil, fmt.Errorf("repository: find user by id: %w", err)
	}

	return &u, nil
}

// FindAll returns all users.
func (r *mysqlUserRepo) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `SELECT id, name FROM user`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("repository: scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate user rows: %w", err)
	}

	return users, nil
}

// DeleteByID deletes the user with the given primary key.
//
// MIGRATION_NOTE: Returns ErrEmptyResultDelete when RowsAffected == 0 to
// replicate Spring Data's EmptyResultDataAccessException (Change 12). This
// error is intentionally NOT a NotFoundError, preserving the original
// unhandled-500 behaviour on delete-of-missing-id.
func (r *mysqlUserRepo) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM user WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user by id: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user by id: read rows affected: %w", err)
	}
	if affected == 0 {
		return ErrEmptyResultDelete
	}

	return nil
}

// FindByName returns the user with the given name.
//
// MIGRATION_NOTE: This is the Go equivalent of Spring Data's derived query
// method findByName(String). The name-to-SQL translation is now explicit. A
// missing row maps to ErrUserNotFound (the Java method returned null).
func (r *mysqlUserRepo) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `SELECT id, name FROM user WHERE name = ?`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, name).Scan(&u.ID, &u.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, smartErr.NewUserNotFound(fmt.Sprintf("user with name %q not found", name))
	case err != nil:
		return nil, fmt.Errorf("repository: find user by name: %w", err)
	}

	return &u, nil
}
