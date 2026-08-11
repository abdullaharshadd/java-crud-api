// Package repository provides data-access logic for the Smart Contact
// service. It is the Go equivalent of the source project's
// com.smartContact.repository package.
//
// MIGRATION_NOTE: The Java source was a Spring Data JPA repository interface
// (UserDao) extending JpaRepository<User, Integer>. Spring generated a proxy
// implementation at runtime, deriving SQL from method names (e.g. findByName ->
// WHERE name = ?) and supplying standard CRUD operations (save, findById,
// findAll, deleteById, ...) for free.
//
// Go has no runtime proxy generation, no annotation-driven query derivation,
// and no ORM in the standard library. The idiomatic replacement is:
//
//   - A UserRepository interface declaring exactly the operations this service
//     uses (the subset of JpaRepository we actually need, plus the custom
//     FindByName finder). Callers depend on the interface, not the concrete
//     type, which keeps the service layer testable.
//   - A concrete PostgresUserRepository backed by *sql.DB that writes the SQL
//     explicitly.
//
// Translation decisions:
//
//   - JPA save() maps to a PostgreSQL upsert (INSERT ... ON CONFLICT ... DO
//     UPDATE) using RETURNING id to obtain the generated key. There is no
//     LastInsertId() with Postgres drivers.
//   - findById / findByName raise UserNotFoundError via NewUserNotFoundErrorf
//     with the exact interpolated messages when no row matches (CHANGE 18).
//   - Every I/O method takes context.Context as its first parameter.
//   - PostgreSQL positional placeholders ($1, $2, ...) are used throughout.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartErr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the persistence operations the Smart Contact service
// needs for the User entity. It is the Go equivalent of the Spring Data JPA
// UserDao interface, narrowed to the operations actually used by callers.
type UserRepository interface {
	// Save persists the given user. If the user has a zero ID a new row is
	// inserted; otherwise the existing row is updated. The persisted user
	// (including any generated ID) is returned.
	Save(ctx context.Context, u *model.User) (*model.User, error)

	// FindByID returns the user with the given ID, or a UserNotFoundError if
	// no such user exists.
	FindByID(ctx context.Context, id int) (*model.User, error)

	// FindByName returns the user with the given name, or a UserNotFoundError
	// if no such user exists. This mirrors the derived findByName query.
	FindByName(ctx context.Context, name string) (*model.User, error)

	// FindAll returns every user.
	FindAll(ctx context.Context) ([]*model.User, error)

	// DeleteByID removes the user with the given ID. It returns a
	// UserNotFoundError if no such user exists.
	DeleteByID(ctx context.Context, id int) error
}

// PostgresUserRepository is a PostgreSQL-backed implementation of
// UserRepository built on the standard database/sql package.
type PostgresUserRepository struct {
	db *sql.DB
}

// compile-time assertion that PostgresUserRepository satisfies UserRepository.
var _ UserRepository = (*PostgresUserRepository)(nil)

// NewPostgresUserRepository constructs a PostgresUserRepository backed by the
// given *sql.DB. It returns an error if db is nil.
func NewPostgresUserRepository(db *sql.DB) (*PostgresUserRepository, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	return &PostgresUserRepository{db: db}, nil
}

// Save persists the given user. When u.ID is zero a new row is inserted and the
// generated ID is returned; otherwise the existing row is updated.
//
// MIGRATION_NOTE: JPA save() performs an INSERT-or-UPDATE based on whether the
// entity is transient. We express that with an explicit branch plus RETURNING
// id (the Postgres way to obtain a generated key).
func (r *PostgresUserRepository) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("repository: user must not be nil")
	}
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("repository: invalid user: %w", err)
	}

	if u.ID == 0 {
		const insertQuery = `
			INSERT INTO users (name, password)
			VALUES ($1, $2)
			RETURNING id`
		if err := r.db.QueryRowContext(ctx, insertQuery, u.Name, u.Password).Scan(&u.ID); err != nil {
			return nil, fmt.Errorf("repository: insert user: %w", err)
		}
		return u, nil
	}

	const updateQuery = `
		UPDATE users
		SET name = $1, password = $2
		WHERE id = $3`
	res, err := r.db.ExecContext(ctx, updateQuery, u.Name, u.Password, u.ID)
	if err != nil {
		return nil, fmt.Errorf("repository: update user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository: update user rows affected: %w", err)
	}
	if n == 0 {
		return nil, smartErr.NewUserNotFoundErrorf("User not found with id %d", u.ID)
	}
	return u, nil
}

// FindByID returns the user with the given ID, or a UserNotFoundError when no
// matching row exists.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `
		SELECT id, name, password
		FROM users
		WHERE id = $1`
	u, err := r.scanUser(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, smartErr.NewUserNotFoundErrorf("User not found with id %d", id)
		}
		return nil, fmt.Errorf("repository: find user by id: %w", err)
	}
	return u, nil
}

// FindByName returns the user with the given name, or a UserNotFoundError when
// no matching row exists. It is the Go equivalent of the derived findByName
// query.
func (r *PostgresUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `
		SELECT id, name, password
		FROM users
		WHERE name = $1`
	u, err := r.scanUser(r.db.QueryRowContext(ctx, query, name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, smartErr.NewUserNotFoundErrorf("User not found with name %s", name)
		}
		return nil, fmt.Errorf("repository: find user by name: %w", err)
	}
	return u, nil
}

// FindAll returns every user in the table.
func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	const query = `
		SELECT id, name, password
		FROM users
		ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate users: %w", err)
	}
	return users, nil
}

// DeleteByID removes the user with the given ID. It returns a
// UserNotFoundError when no matching row exists.
func (r *PostgresUserRepository) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user by id: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user rows affected: %w", err)
	}
	if n == 0 {
		return smartErr.NewUserNotFoundErrorf("User not found with id %d", id)
	}
	return nil
}

// rowScanner abstracts *sql.Row and *sql.Rows so scanUser can serve both
// single-row and multi-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser reads a single user row from the given scanner.
func (r *PostgresUserRepository) scanUser(s rowScanner) (*model.User, error) {
	var u model.User
	if err := s.Scan(&u.ID, &u.Name, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}
