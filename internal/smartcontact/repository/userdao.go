// Package repository provides data-access implementations for the
// smartContact application. It replaces the Spring Data JPA repository
// interfaces with explicit database/sql-backed code.
//
// MIGRATION_NOTE: The Java source (UserDao) was a Spring Data JPA repository
// interface that extended JpaRepository<User, Integer> and declared a single
// derived query method, findByName. Spring generated the implementation at
// runtime for both the inherited CRUD methods (findAll, findById, save, etc.)
// and the derived finder. Go has no runtime proxy generation, so this file
// provides a concrete UserDao backed by *sql.DB, implementing the CRUD
// operations that the rest of the application actually uses plus the custom
// FindByName finder.
//
// MIGRATION_NOTE: All read methods scan into an internal userRow using
// sql.NullString to tolerate NULL columns (which previously caused
// "converting NULL to string" 500 errors), then convert to model.UserResponse
// via toResponse. Merge (the save/upsert) reuses the same userRow/toResponse
// path for its post-write re-read rather than duplicating a converter.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
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

// userRow is the internal scan target for user reads. Nullable text columns
// use sql.NullString so that NULL values do not fail the scan.
//
// MIGRATION_NOTE: Table/column names follow the PostgreSQL lower_snake_case
// convention established in model.User (users.id, name, email, etc.), so no
// identifier quoting is required.
type userRow struct {
	id       int
	name     sql.NullString
	second   sql.NullString
	work     sql.NullString
	email    sql.NullString
	phone    sql.NullString
	image    sql.NullString
	description sql.NullString
	userName sql.NullString
	password sql.NullString
}

// selectColumns lists the columns read by every finder, in scan order.
const selectColumns = "id, name, second_name, work, email, phone_number, image, description, user_name, password"

// scan reads one row into a userRow.
func (r *userRow) scan(row interface{ Scan(...any) error }) error {
	return row.Scan(
		&r.id,
		&r.name,
		&r.second,
		&r.work,
		&r.email,
		&r.phone,
		&r.image,
		&r.description,
		&r.userName,
		&r.password,
	)
}

// toResponse converts an internal userRow into the shared wire shape
// model.UserResponse, unwrapping sql.NullString into plain strings.
func (r *userRow) toResponse() model.UserResponse {
	return model.UserResponse{
		ID:          r.id,
		Name:        r.name.String,
		SecondName:  r.second.String,
		Work:        r.work.String,
		Email:       r.email.String,
		PhoneNumber: r.phone.String,
		Image:       r.image.String,
		Description: r.description.String,
		UserName:    r.userName.String,
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
		var r userRow
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

	var r userRow
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

	var r userRow
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
// re-reads and returns the persisted row via the shared userRow/toResponse
// path.
//
// MIGRATION_NOTE: Replaces JpaRepository.save(entity), which performs an
// insert-or-update based on the presence of the identifier. Inserts use
// PostgreSQL RETURNING id to obtain the generated primary key (there is no
// LastInsertId equivalent for Postgres drivers).
func (d *UserDao) Merge(ctx context.Context, u model.User) (model.UserResponse, error) {
	if u.ID == 0 {
		const insert = `INSERT INTO users
			(name, second_name, work, email, phone_number, image, description, user_name, password)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`

		var id int
		err := d.db.QueryRowContext(ctx, insert,
			u.Name, u.SecondName, u.Work, u.Email, u.PhoneNumber,
			u.Image, u.Description, u.UserName, u.Password,
		).Scan(&id)
		if err != nil {
			return model.UserResponse{}, fmt.Errorf("repository: insert user: %w", err)
		}
		return d.FindByID(ctx, id)
	}

	const update = `UPDATE users SET
		name = $1,
		second_name = $2,
		work = $3,
		email = $4,
		phone_number = $5,
		image = $6,
		description = $7,
		user_name = $8,
		password = $9
		WHERE id = $10`

	res, err := d.db.ExecContext(ctx, update,
		u.Name, u.SecondName, u.Work, u.Email, u.PhoneNumber,
		u.Image, u.Description, u.UserName, u.Password, u.ID,
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
