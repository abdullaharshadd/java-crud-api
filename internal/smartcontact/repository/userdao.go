// Package repository provides data-access implementations for the smartcontact
// service. Each type in this package wraps a database handle and exposes
// explicit, context-aware methods for persisting and retrieving domain models.
//
// MIGRATION_NOTE: The Java source was a Spring Data JPA repository interface
// (UserDao extends JpaRepository<User, Integer>). Spring auto-implements such
// interfaces at runtime, deriving SQL from method names (e.g. findByName) and
// providing a full CRUD surface (save, findById, findAll, deleteById, ...).
//
// Go has no runtime proxying or query derivation, so this is translated into a
// concrete UserRepository backed by sqlx. The standard CRUD operations that
// JpaRepository would have supplied for free are written out explicitly here,
// plus the single custom finder (FindByName). All I/O methods take a
// context.Context and return explicit errors.
//
// MIGRATION_NOTE (blocked on V-SCHEMA): The exact column names, id generation
// strategy, and per-column nullability derive from the JPA @Entity/@Column
// annotations which are not fully resolved here. The column list below uses the
// documented mapping from model/user.go, lower_snake_cased for PostgreSQL. If a
// V-SCHEMA artifact later disagrees, the column constants must be reconciled.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	apperror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// Querier is the minimal database surface the UserRepository depends on. Both
// *sqlx.DB and *sqlx.Tx satisfy it, so callers may run repository operations
// either standalone or inside an existing transaction.
//
// MIGRATION_NOTE: This interface stands in for JpaRepository's implicit
// EntityManager. Depending on an interface (rather than a concrete *sqlx.DB)
// keeps the repository testable and transaction-agnostic.
type Querier interface {
	sqlx.ExtContext
}

// UserRepository provides persistence operations for model.User.
//
// It replaces the Spring Data UserDao interface. Unlike the auto-implemented
// JPA repository, every method here is explicit and context-aware.
type UserRepository struct {
	db Querier
}

// NewUserRepository constructs a UserRepository backed by the given Querier.
//
// The Querier may be a *sqlx.DB for standalone use or a *sqlx.Tx to enroll the
// repository in an existing transaction.
func NewUserRepository(db Querier) *UserRepository {
	return &UserRepository{db: db}
}

// userColumns lists the persisted, non-generated columns for the user table in
// insert/update order.
//
// MIGRATION_NOTE (blocked on V-SCHEMA): names lower_snake_cased from the Java
// @Column mapping documented in model/user.go.
const (
	userTable = "users"
)

// FindByName looks up a single user by their name.
//
// This is the direct translation of the Spring Data derived query method
// User findByName(String name). In JPA a missing row surfaced as null (or
// threw depending on configuration); here a missing row is reported explicitly
// as apperror.ErrUserNotFound so callers can branch with
// apperror.IsUserNotFound.
func (r *UserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `SELECT id, name, email, password, role, about FROM ` + userTable + ` WHERE name = $1`

	var u model.User
	if err := sqlx.GetContext(ctx, r.db, &u, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewUserNotFoundErrorf("user with name %q not found", name)
		}
		return nil, fmt.Errorf("find user by name %q: %w", name, err)
	}
	return &u, nil
}

// FindByID retrieves a user by primary key.
//
// MIGRATION_NOTE: Replaces JpaRepository.findById(Integer). JPA returned an
// Optional<User>; the idiomatic Go equivalent is (*model.User, error) where a
// missing row yields apperror.ErrUserNotFound.
func (r *UserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `SELECT id, name, email, password, role, about FROM ` + userTable + ` WHERE id = $1`

	var u model.User
	if err := sqlx.GetContext(ctx, r.db, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewUserNotFoundErrorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("find user by id %d: %w", id, err)
	}
	return &u, nil
}

// FindAll returns every user in the table.
//
// MIGRATION_NOTE: Replaces JpaRepository.findAll(). Returns a slice rather than
// a Java List; an empty result set yields an empty (non-nil) slice.
func (r *UserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `SELECT id, name, email, password, role, about FROM ` + userTable + ` ORDER BY id`

	users := []model.User{}
	if err := sqlx.SelectContext(ctx, r.db, &users, query); err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	return users, nil
}

// Save persists a user, faithfully replicating JPA's save(entity) merge
// semantics keyed on the primary key only.
//
// MIGRATION_NOTE: Spring Data's save() performs an insert when the entity has
// no id and a merge (update) when it does. Per the migration debate this is
// modelled with an explicit exists-then-branch keyed on the PK, NOT a Postgres
// ON CONFLICT upsert, to preserve the exact JPA behaviour (merge keys on the
// primary key, not on any unique column such as email).
//
// MIGRATION_NOTE (blocked on V-SCHEMA): the id-generation path assumes an
// identity/serial primary key. When u.ID == 0 the row is inserted and the
// generated id is returned via RETURNING id (Postgres has no LastInsertId).
func (r *UserRepository) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("save user: user is nil")
	}

	if u.ID != 0 {
		exists, err := r.existsByID(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return r.update(ctx, u)
		}
	}
	return r.insert(ctx, u)
}

// existsByID reports whether a user row with the given id is present.
func (r *UserRepository) existsByID(ctx context.Context, id int) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM ` + userTable + ` WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRowxContext(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists by id %d: %w", id, err)
	}
	return exists, nil
}

// insert adds a new user row and populates the generated id on the returned copy.
func (r *UserRepository) insert(ctx context.Context, u *model.User) (*model.User, error) {
	const query = `
		INSERT INTO ` + userTable + ` (name, email, password, role, about)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	saved := *u
	if err := r.db.QueryRowxContext(ctx, query,
		u.Name, u.Email, u.Password, u.Role, u.About,
	).Scan(&saved.ID); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &saved, nil
}

// update rewrites an existing user row identified by its primary key.
func (r *UserRepository) update(ctx context.Context, u *model.User) (*model.User, error) {
	const query = `
		UPDATE ` + userTable + `
		SET name = $1, email = $2, password = $3, role = $4, about = $5
		WHERE id = $6`

	res, err := r.db.ExecContext(ctx, query,
		u.Name, u.Email, u.Password, u.Role, u.About, u.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update user id %d: %w", u.ID, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update user id %d: rows affected: %w", u.ID, err)
	}
	if n == 0 {
		return nil, apperror.NewUserNotFoundErrorf("user with id %d not found", u.ID)
	}

	saved := *u
	return &saved, nil
}

// DeleteByID removes a user by primary key.
//
// MIGRATION_NOTE: Replaces JpaRepository.deleteById(Integer). JPA threw
// EmptyResultDataAccessException when the id was absent; here an absent id
// yields apperror.ErrUserNotFound.
func (r *UserRepository) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM ` + userTable + ` WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user id %d: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user id %d: rows affected: %w", id, err)
	}
	if n == 0 {
		return apperror.NewUserNotFoundErrorf("user with id %d not found", id)
	}
	return nil
}
