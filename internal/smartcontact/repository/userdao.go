package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the persistence operations available for users.
//
// MIGRATION_NOTE: This interface captures the subset of JpaRepository behavior
// the application relies on, plus the derived FindByName finder from the Java
// UserDao. Depending on an interface (rather than the concrete *UserDao) keeps
// callers testable and decoupled from database/sql.
type UserRepository interface {
	// FindAll returns every user in the users table.
	FindAll(ctx context.Context) ([]model.UserResponse, error)
	// FindByID returns the user with the given id, or ErrUserNotFound.
	FindByID(ctx context.Context, id int) (model.UserResponse, error)
	// FindByName returns the user with the given name, or ErrUserNotFound.
	FindByName(ctx context.Context, name string) (model.UserResponse, error)
	// Merge inserts a new user (when ID == 0) or updates an existing one,
	// then re-reads and returns the persisted row.
	Merge(ctx context.Context, u model.User) (model.UserResponse, error)
	// DeleteByID removes the user with the given id.
	DeleteByID(ctx context.Context, id int) error
}

// UserDao is the database/sql-backed implementation of UserRepository.
//
// MIGRATION_NOTE: The name UserDao is preserved from the Java source, but here
// it is a concrete struct rather than an interface (the interface role is
// played by UserRepository above).
type UserDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserDao backed by the given *sql.DB. It replaces
// Spring's runtime-generated repository proxy and component scanning.
func NewUserDao(db *sql.DB) *UserDao {
	return &UserDao{db: db}
}

// compile-time assertion that *UserDao satisfies UserRepository.
var _ UserRepository = (*UserDao)(nil)

// localUserRow is the internal scan target for user reads. Nullable text columns
// use sql.NullString so that NULL values do not fail the scan.
//
// MIGRATION_NOTE: Table/column names follow the PostgreSQL lower_snake_case
// convention established in model.User (users.id, name, email, etc.), so no
// identifier quoting is required.
type localUserRow struct {
	id    int
	name  sql.NullString
	email sql.NullString
	role  sql.NullString
	about sql.NullString
}

// selectColumns lists the columns read by every finder, in scan order.
const selectColumns = "id, name, email, role, about"

// scan reads one row into a localUserRow.
func (r *localUserRow) scan(row interface{ Scan(...any) error }) error {
	return row.Scan(
		&r.id,
		&r.name,
		&r.email,
		&r.role,
		&r.about,
	)
}

// toResponse converts an internal localUserRow into the shared wire shape
// model.UserResponse, unwrapping sql.NullString into plain *string.
func (r *localUserRow) toResponse() model.UserResponse {
	nsToPtr := func(ns sql.NullString) *string {
		if !ns.Valid {
			return nil
		}
		s := ns.String
		return &s
	}
	return model.UserResponse{
		ID:    r.id,
		Name:  nsToPtr(r.name),
		Email: nsToPtr(r.email),
		Role:  nsToPtr(r.role),
		About: nsToPtr(r.about),
	}
}

// FindAll returns every user in the users table.
//
// MIGRATION_NOTE: Replaces JpaRepository.findAll().
func (d *UserDao) FindAll(ctx context.Context) ([]model.UserResponse, error) {
	const query = "SELECT " + selectColumns + " FROM users ORDER BY id"

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var r localUserRow
		if err := r.scan(rows); err != nil {
			return nil, fmt.Errorf("repository: scan user row: %w", err)
		}
		users = append(users, r.toResponse())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate user rows: %w", err)
	}
	return users, nil
}

// FindByID returns the user with the given id, or a UserNotFoundError wrapping
// ErrUserNotFound when no such user exists.
//
// MIGRATION_NOTE: Replaces JpaRepository.findById(Integer). Java returned an
// Optional<User>; here absence is signalled by ErrUserNotFound rather than an
// empty Optional.
func (d *UserDao) FindByID(ctx context.Context, id int) (model.UserResponse, error) {
	const query = "SELECT " + selectColumns + " FROM users WHERE id = $1"

	var r localUserRow
	if err := r.scan(d.db.QueryRowContext(ctx, query, id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.UserResponse{}, smartcontacterror.NewUserNotFoundError(
				fmt.Sprintf("user with id %d not found", id),
			)
		}
		return model.UserResponse{}, fmt.Errorf("repository: find user by id %d: %w", id, err)
	}
	return r.toResponse(), nil
}

// FindByName returns the user with the given name, or a UserNotFoundError
// wrapping ErrUserNotFound when no such user exists.
//
// MIGRATION_NOTE: This is the direct translation of the Java derived query
// method `User findByName(String name)`. Spring derived the WHERE clause from
// the method name; here it is written explicitly with a PostgreSQL positional
// placeholder.
func (d *UserDao) FindByName(ctx context.Context, name string) (model.UserResponse, error) {
	const query = "SELECT " + selectColumns + " FROM users WHERE name = $1"

	var r localUserRow
	if err := r.scan(d.db.QueryRowContext(ctx, query, name)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.UserResponse{}, smartcontacterror.NewUserNotFoundError(
				fmt.Sprintf("user with name %q not found", name),
			)
		}
		return model.UserResponse{}, fmt.Errorf("repository: find user by name %q: %w", name, err)
	}
	return r.toResponse(), nil
}

// Merge inserts a new user (when u.ID == 0) or updates an existing one, then
// re-reads and returns the persisted row via the shared localUserRow/toResponse
// path.
//
// MIGRATION_NOTE: Replaces JpaRepository.save(entity), which performs an
// insert-or-update based on the presence of the identifier. Inserts use
// PostgreSQL RETURNING id to obtain the generated primary key (there is no
// LastInsertId equivalent for Postgres drivers).
func (d *UserDao) Merge(ctx context.Context, u model.User) (model.UserResponse, error) {
	if u.ID == 0 {
		const insert = `INSERT INTO users
			(name, email, password, role, about)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`

		var id int
		err := d.db.QueryRowContext(ctx, insert,
			u.Name, u.Email, u.Password, u.Role, u.About,
		).Scan(&id)
		if err != nil {
			return model.UserResponse{}, fmt.Errorf("repository: insert user: %w", err)
		}
		return d.FindByID(ctx, id)
	}

	const update = `UPDATE users SET
		name = $1,
		email = $2,
		password = $3,
		role = $4,
		about = $5
		WHERE id = $6`

	res, err := d.db.ExecContext(ctx, update,
		u.Name, u.Email, u.Password, u.Role, u.About, u.ID,
	)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("repository: update user %d: %w", u.ID, err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return model.UserResponse{}, smartcontacterror.NewUserNotFoundError(
			fmt.Sprintf("user with id %d not found", u.ID),
		)
	}
	return d.FindByID(ctx, u.ID)
}

// DeleteByID removes the user with the given id, returning a UserNotFoundError
// when no matching row exists.
//
// MIGRATION_NOTE: Replaces JpaRepository.deleteById(Integer).
func (d *UserDao) DeleteByID(ctx context.Context, id int) error {
	const query = "DELETE FROM users WHERE id = $1"

	res, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user %d: rows affected: %w", id, err)
	}
	if n == 0 {
		return smartcontacterror.NewUserNotFoundError(
			fmt.Sprintf("user with id %d not found", id),
		)
	}
	return nil
}