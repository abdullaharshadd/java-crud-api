// Package repository provides data-access implementations for the
// smartcontact domain types.
//
// MIGRATION_NOTE: The original Java type was a Spring Data JPA repository
// interface:
//
//	@Repository
//	public interface UserDao extends JpaRepository<User,Integer> {
//	    public User findByName(String name);
//	}
//
// Spring Data generated a full CRUD implementation at runtime purely from the
// interface declaration, plus a derived query (findByName) from the method
// name. Go has no such runtime code generation, so the CRUD operations and the
// custom finder are implemented explicitly here against a *sqlx.DB.
//
// The JpaRepository<User, Integer> contract is reproduced by the methods on
// UserRepository (Save, FindByID, FindAll, DeleteByID) plus the derived
// FindByName query.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// userColumns is the explicit, aliased column list used for all SELECT
// statements against the USER table. Using explicit AS aliases keeps the
// query resilient to changes in the model's db tags and avoids relying on
// SELECT * ordering.
//
// MIGRATION_NOTE: Column names below are assumed to match the User model's
// `db` tags. Verify against the actual schema during review.
const userColumns = `id AS id, name AS name, secondName AS secondName, work AS work, email AS email, phone AS phone, about AS about`

// UserRepository defines the persistence operations for the User entity.
//
// MIGRATION_NOTE: This interface stands in for Spring's JpaRepository<User,
// Integer>. Only the operations actually needed by the application are
// exposed; the full JpaRepository surface (paging, sorting, batch ops) is
// intentionally omitted and can be added on demand.
type UserRepository interface {
	// Save inserts a new user or updates an existing one, returning the
	// persisted user (including any generated ID).
	Save(ctx context.Context, user model.User) (model.User, error)
	// FindByID returns the user with the given ID, or a UserNotFoundError if
	// no such user exists.
	FindByID(ctx context.Context, id int) (model.User, error)
	// FindAll returns every user in the table.
	FindAll(ctx context.Context) ([]model.User, error)
	// FindByName returns the first user matching the given name.
	FindByName(ctx context.Context, name string) (model.User, error)
	// DeleteByID removes the user with the given ID.
	DeleteByID(ctx context.Context, id int) error
}

// userRepo is the sqlx-backed implementation of UserRepository.
type userRepo struct {
	db *sqlx.DB
}

// NewUserRepository constructs a UserRepository backed by the given database
// handle.
//
// MIGRATION_NOTE: Dependency injection is done explicitly via this constructor
// rather than through Spring's @Repository component scanning.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepo{db: db}
}

// Save inserts a new user (when ID is zero) or updates an existing one, then
// returns the persisted user.
func (r *userRepo) Save(ctx context.Context, user model.User) (model.User, error) {
	if err := user.Validate(); err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}

	if user.ID == 0 {
		return r.insert(ctx, user)
	}
	return r.update(ctx, user)
}

func (r *userRepo) insert(ctx context.Context, user model.User) (model.User, error) {
	const query = `INSERT INTO USER (name, secondName, work, email, phone, about)
		VALUES (:name, :secondName, :work, :email, :phone, :about)`

	res, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.User{}, fmt.Errorf("insert user: read generated id: %w", err)
	}
	user.ID = int(id)
	return user, nil
}

func (r *userRepo) update(ctx context.Context, user model.User) (model.User, error) {
	const query = `UPDATE USER SET name = :name, secondName = :secondName,
		work = :work, email = :email, phone = :phone, about = :about
		WHERE id = :id`

	res, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}

	if _, err := res.RowsAffected(); err != nil {
		return model.User{}, fmt.Errorf("update user: read rows affected: %w", err)
	}
	return user, nil
}

// FindByID returns the user with the given ID, or a UserNotFoundError if the
// row does not exist.
func (r *userRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	query := `SELECT ` + userColumns + ` FROM USER WHERE id = ?`

	var user model.User
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, smartError.NewUserNotFoundError(id)
		}
		return model.User{}, fmt.Errorf("find user by id %d: %w", id, err)
	}
	return user, nil
}

// FindAll returns all users in the USER table.
func (r *userRepo) FindAll(ctx context.Context) ([]model.User, error) {
	query := `SELECT ` + userColumns + ` FROM USER`

	var users []model.User
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	return users, nil
}

// FindByName returns the first user matching the given name.
//
// MIGRATION_NOTE: This reproduces Spring Data's derived query `findByName`.
// The original Java method returned a single User (null if none found). Here a
// missing row is reported explicitly rather than via a nil/zero value, and
// only the first matching row is returned (LIMIT 1) since the original return
// type was a single User.
func (r *userRepo) FindByName(ctx context.Context, name string) (model.User, error) {
	query := `SELECT ` + userColumns + ` FROM USER WHERE name = ? LIMIT 1`

	var user model.User
	if err := r.db.GetContext(ctx, &user, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("find user by name %q: %w", name, sql.ErrNoRows)
		}
		return model.User{}, fmt.Errorf("find user by name %q: %w", name, err)
	}
	return user, nil
}

// DeleteByID removes the user with the given ID.
//
// MIGRATION_NOTE: Spring Data's deleteById throws an
// EmptyResultDataAccessException (surfacing as an HTTP 500) when the target
// ID does not exist. To replicate that behaviour, this method inspects
// RowsAffected() and returns a non-NotFound error when no row was deleted,
// rather than a UserNotFoundError (which would map to a 404).
func (r *userRepo) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM USER WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user by id %d: %w", id, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user by id %d: read rows affected: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("delete user by id %d: no rows affected", id)
	}
	return nil
}
