// Package repository provides data-access objects (DAOs) for the SmartContact
// service. It replaces the Java Spring Data JPA repository layer under
// com.smartContact.repository.
//
// MIGRATION_NOTE: The Java UserDao was a Spring Data JPA repository interface:
//
//	@Repository
//	public interface UserDao extends JpaRepository<User,Integer> {
//	    public User findByName(String name);
//	}
//
// Spring generated the entire implementation at runtime: CRUD methods from
// JpaRepository (save/findById/findAll/deleteById/etc.) plus a derived query
// for findByName. Idiomatic Go has no such code generation, so the equivalent
// is an explicit interface describing the operations this application actually
// uses, backed by a concrete sqlx implementation with hand-written SQL.
//
// Dialect: the target database is PostgreSQL. Placeholders use $1, $2, ...;
// inserts return the generated identity via `RETURNING id` (there is no
// LastInsertId() for the Postgres drivers).
//
// Column/table names follow the lower_snake_case mapping established in
// model.User's db tags (user_id, user_name, user_email, user_password,
// user_role, user_about) — there is no "phone" column or field on model.User;
// an earlier draft of this file referenced one that never matched the actual
// model and never compiled.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// UserDao describes the persistence operations available for the User entity.
//
// MIGRATION_NOTE: In the Java source this surface came "for free" by extending
// JpaRepository<User,Integer>. Here we declare only the operations the
// application needs, expressed as an interface so that callers depend on an
// abstraction (easy to mock in tests). The primary-key type is int (Java's
// Integer). All methods take a context.Context because they perform I/O.
type UserDao interface {
	// Save inserts a new user and returns it populated with the
	// database-generated identifier.
	Save(ctx context.Context, u *model.User) (*model.User, error)
	// Update persists changes to an existing user, identified by its ID.
	// It returns ErrUserNotFound if no row was affected.
	Update(ctx context.Context, u *model.User) error
	// FindByID returns the user with the given identifier, or ErrUserNotFound
	// if none exists.
	FindByID(ctx context.Context, id int) (*model.User, error)
	// FindAll returns every user, ordered by id.
	FindAll(ctx context.Context) ([]model.User, error)
	// FindByName returns the user with the given name. It replaces the derived
	// query findByName(String) from the Spring repository. It returns
	// ErrUserNotFound if none exists.
	FindByName(ctx context.Context, name string) (*model.User, error)
	// DeleteByID removes the user with the given identifier. It returns
	// ErrUserNotFound if no row was affected.
	DeleteByID(ctx context.Context, id int) error
}

// userDao is the sqlx-backed implementation of UserDao.
type userDao struct {
	db *sqlx.DB
}

// NewUserDao constructs a UserDao backed by the given database handle.
//
// MIGRATION_NOTE: This constructor replaces Spring's @Repository component
// scanning and dependency injection. Callers wire the *sqlx.DB explicitly.
func NewUserDao(db *sqlx.DB) UserDao {
	return &userDao{db: db}
}

// Save inserts a new user and returns it with the generated id populated.
//
// MIGRATION_NOTE: JpaRepository.save both inserts and updates depending on the
// presence of an identifier. Here Save is insert-only (RETURNING id); use
// Update for modifications. This keeps the SQL explicit and avoids surprising
// upsert semantics.
func (d *userDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("repository: cannot save nil user")
	}

	const query = `
		INSERT INTO users (user_name, user_email, user_password, user_role, user_about)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id`

	var id int
	if err := d.db.QueryRowxContext(
		ctx, query, u.Name, u.Email, u.Password, u.Role, u.About,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("repository: save user: %w", err)
	}
	u.ID = id
	return u, nil
}

// Update persists changes to an existing user, matched by ID.
func (d *userDao) Update(ctx context.Context, u *model.User) error {
	if u == nil {
		return errors.New("repository: cannot update nil user")
	}

	const query = `
		UPDATE users
		SET user_name = $1, user_email = $2, user_password = $3, user_role = $4, user_about = $5
		WHERE user_id = $6`

	res, err := d.db.ExecContext(
		ctx, query, u.Name, u.Email, u.Password, u.Role, u.About, u.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: update user %d: %w", u.ID, err)
	}

	// Affected-rows pattern: a zero-row update means the id did not exist.
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: update user %d: rows affected: %w", u.ID, err)
	}
	if affected == 0 {
		return smartcontacterror.ErrUserNotFound
	}
	return nil
}

// FindByID returns the user with the given identifier.
func (d *userDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `
		SELECT user_id, user_name, user_email, user_password, user_role, user_about
		FROM users
		WHERE user_id = $1`

	var u model.User
	if err := d.db.GetContext(ctx, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, smartcontacterror.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: find user %d: %w", id, err)
	}
	return &u, nil
}

// FindAll returns every user, ordered by id for stable output.
func (d *userDao) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `
		SELECT user_id, user_name, user_email, user_password, user_role, user_about
		FROM users
		ORDER BY user_id`

	var users []model.User
	if err := d.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	return users, nil
}

// FindByName returns the user with the given name.
//
// MIGRATION_NOTE: This is the direct equivalent of the Spring derived query
// findByName(String). Spring returned a single User (or null); Go returns
// ErrUserNotFound when there is no match. If name is not unique in the real
// schema, this query would return an arbitrary row without an ORDER BY — the
// ordering key MUST be confirmed against the actual schema during review. An
// explicit `ORDER BY user_id LIMIT 1` is applied so behaviour is deterministic.
func (d *userDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `
		SELECT user_id, user_name, user_email, user_password, user_role, user_about
		FROM users
		WHERE user_name = $1
		ORDER BY user_id
		LIMIT 1`

	var u model.User
	if err := d.db.GetContext(ctx, &u, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, smartcontacterror.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: find user by name %q: %w", name, err)
	}
	return &u, nil
}

// DeleteByID removes the user with the given identifier.
func (d *userDao) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE user_id = $1`

	res, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user %d: %w", id, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user %d: rows affected: %w", id, err)
	}
	if affected == 0 {
		return smartcontacterror.ErrUserNotFound
	}
	return nil
}
