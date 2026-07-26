package repository

// Package repository provides data-access objects (DAOs) for the SmartContact
// service. This file corresponds to the original Spring Data JPA interface
// com.smartContact.repository.UserDao.
//
// MIGRATION_NOTE: The Java source was a Spring Data JPA repository interface
// (extends JpaRepository<User, Integer>) annotated with @Repository. Spring
// generated the implementation at runtime, including all CRUD methods (save,
// findById, findAll, deleteById, count, ...) and derived-query finders such as
// findByName. Go has no ORM/proxy-generation equivalent, so the concrete
// implementation is written explicitly here against database/sql. The set of
// methods mirrors the JpaRepository CRUD surface actually needed by the
// application plus the custom FindByName finder.
//
// MIGRATION_NOTE: The original config targeted MySQL, but the migration target
// is PostgreSQL. All queries therefore use $1, $2 positional placeholders and
// INSERT ... RETURNING id for generated keys (Postgres drivers do not support
// LastInsertId()).
package repository

import (
	"context"
	"database/sql"
	errs "errors"
	"fmt"

	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// UserDao defines persistence operations for User entities. It mirrors the
// subset of Spring Data JpaRepository behaviour used by the SmartContact
// service, including the custom FindByName finder.
type UserDao interface {
	// Save persists the given user. If the user has a zero ID it is inserted
	// and the generated ID is returned; otherwise the existing row is updated
	// (upserted). It returns the persisted user.
	Save(ctx context.Context, user *model.User) (*model.User, error)

	// FindByID returns the user with the given ID. It returns a
	// *error.UserNotFoundError if no such user exists.
	FindByID(ctx context.Context, id int) (*model.User, error)

	// FindAll returns every user. It always returns a non-nil slice so an
	// empty table serializes as [] rather than null.
	FindAll(ctx context.Context) ([]model.User, error)

	// FindByName returns the user with the given name. It returns a
	// *error.UserNotFoundError if no such user exists.
	FindByName(ctx context.Context, name string) (*model.User, error)

	// DeleteByID removes the user with the given ID. It returns a
	// *error.UserNotFoundError if no such user exists.
	DeleteByID(ctx context.Context, id int) error

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)
}

// userDao is the database/sql-backed implementation of UserDao.
type userDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserDao backed by the given *sql.DB.
func NewUserDao(db *sql.DB) UserDao {
	return &userDao{db: db}
}

// scanUser scans a single user row (in id, name column order) from the given
// scanner into a *model.User.
func scanUser(scanner interface {
	Scan(dest ...any) error
}) (*model.User, error) {
	var (
		id   int
		name string
	)
	if err := scanner.Scan(&id, &name); err != nil {
		return nil, err
	}
	user := model.NewUser(name)
	user.ID = id
	return user, nil
}

// Save persists the given user, inserting when the ID is zero and upserting
// otherwise. Each path is a single atomic statement to preserve auto-commit
// semantics.
func (d *userDao) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, errs.New("repository: cannot save nil user")
	}

	if user.ID == 0 {
		const query = `INSERT INTO users (name) VALUES ($1) RETURNING id`
		if err := d.db.QueryRowContext(ctx, query, user.Name).Scan(&user.ID); err != nil {
			return nil, fmt.Errorf("repository: insert user: %w", err)
		}
		return user, nil
	}

	const query = `INSERT INTO users (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`
	if _, err := d.db.ExecContext(ctx, query, user.ID, user.Name); err != nil {
		return nil, fmt.Errorf("repository: upsert user id=%d: %w", user.ID, err)
	}
	return user, nil
}

// FindByID returns the user with the given ID.
func (d *userDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE id = $1`
	user, err := scanUser(d.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errs.Is(err, sql.ErrNoRows) {
			return nil, smartErr.NewUserNotFoundError(fmt.Sprintf("user not found with id: %d", id))
		}
		return nil, fmt.Errorf("repository: find user by id=%d: %w", id, err)
	}
	return user, nil
}

// FindAll returns every user. The returned slice is always non-nil.
func (d *userDao) FindAll(ctx context.Context) ([]model.User, error) {
	const query = `SELECT id, name FROM users`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	// Initialize a non-nil slice so an empty table serializes as [] not null.
	users := []model.User{}
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("repository: scan user row: %w", scanErr)
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate user rows: %w", err)
	}
	return users, nil
}

// FindByName returns the user with the given name.
func (d *userDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE name = $1`
	user, err := scanUser(d.db.QueryRowContext(ctx, query, name))
	if err != nil {
		if errs.Is(err, sql.ErrNoRows) {
			return nil, smartErr.NewUserNotFoundError(fmt.Sprintf("user not found with name: %s", name))
		}
		return nil, fmt.Errorf("repository: find user by name=%q: %w", name, err)
	}
	return user, nil
}

// DeleteByID removes the user with the given ID.
func (d *userDao) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`
	res, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user id=%d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user id=%d rows affected: %w", id, err)
	}
	if affected == 0 {
		return smartErr.NewUserNotFoundError(fmt.Sprintf("user not found with id: %d", id))
	}
	return nil
}

// Count returns the total number of users.
func (d *userDao) Count(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM users`
	var count int64
	if err := d.db.QueryRowContext(ctx, query, ).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count users: %w", err)
	}
	return count, nil
}
