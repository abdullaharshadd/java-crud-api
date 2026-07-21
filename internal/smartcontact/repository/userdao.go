// Package repository contains the SmartContact service's data-access layer.
// It is the Go equivalent of the Java package com.smartContact.repository.
//
// MIGRATION_NOTE: The source UserDao.java was a Spring Data JPA repository
// interface:
//
//	@Repository
//	public interface UserDao extends JpaRepository<User, Integer> {
//	    public User findByName(String name);
//	}
//
// Spring Data generated an implementation of this interface at runtime via a
// dynamic proxy. JpaRepository<User, Integer> supplied a full set of CRUD
// methods (save, findById, findAll, deleteById, ...) for the User entity
// keyed by an Integer primary key, while findByName was a derived query
// method whose SQL was inferred from the method name.
//
// Go has no equivalent to Spring Data's proxy generation or query derivation.
// The idiomatic analogue is:
//
//   - Define an interface (UserDao) that declares the operations the service
//     actually needs. Callers depend on the interface, not the concrete type
//     (dependency inversion / easy mocking in tests).
//   - Provide a concrete implementation backed by database/sql. Every method
//     takes a context.Context as its first parameter and returns an explicit
//     error, so I/O cancellation and failures are handled explicitly rather
//     than through Hibernate's implicit session management.
//
// The primary key type (Integer in Java) is modelled as Go's int. Note that
// model.User currently exposes no explicit ID/Name accessors in the migrated
// model per the provided context; the SQL below assumes an "users" table with
// "id" and "name" columns and that model.User can be scanned/bound. Adjust the
// column names and scan targets to match the final model.User definition
// during manual review.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	smartcontacterror "internal/smartcontact/error"
	"internal/smartcontact/model"
)

// UserDao describes the persistence operations available for the User entity.
//
// It is the Go analogue of the Spring Data JpaRepository<User, Integer>
// interface. Only the operations the service needs are declared; callers
// depend on this interface so the storage backend can be swapped or mocked.
type UserDao interface {
	// FindByName looks up a single user by name. It returns
	// error.ErrUserNotFound (wrapped) when no matching user exists.
	FindByName(ctx context.Context, name string) (*model.User, error)

	// FindByID looks up a single user by primary key. It returns
	// error.ErrUserNotFound (wrapped) when no matching user exists.
	FindByID(ctx context.Context, id int) (*model.User, error)

	// FindAll returns all persisted users.
	FindAll(ctx context.Context) ([]*model.User, error)

	// Save inserts a new user or updates an existing one and returns the
	// persisted user (including any generated ID).
	Save(ctx context.Context, user *model.User) (*model.User, error)

	// DeleteByID removes the user with the given primary key.
	DeleteByID(ctx context.Context, id int) error
}

// sqlUserDao is a database/sql-backed implementation of UserDao.
type sqlUserDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserDao backed by the given *sql.DB.
//
// MIGRATION_NOTE: This constructor replaces Spring's @Repository
// component-scanning and runtime proxy generation. The *sql.DB is injected
// explicitly rather than autowired.
func NewUserDao(db *sql.DB) (UserDao, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	return &sqlUserDao{db: db}, nil
}

// FindByName looks up a single user by name.
//
// It is the Go equivalent of the derived query method
// User findByName(String name). When no row matches, it returns an error that
// wraps error.ErrUserNotFound so callers can detect it with errors.Is.
func (d *sqlUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE name = ?`

	row := d.db.QueryRowContext(ctx, query, name)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: find user by name %q: %w", name, smartcontacterror.ErrUserNotFound)
		}
		return nil, fmt.Errorf("repository: find user by name %q: %w", name, err)
	}
	return user, nil
}

// FindByID looks up a single user by primary key.
func (d *sqlUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	const query = `SELECT id, name FROM users WHERE id = ?`

	row := d.db.QueryRowContext(ctx, query, id)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: find user by id %d: %w", id, smartcontacterror.ErrUserNotFound)
		}
		return nil, fmt.Errorf("repository: find user by id %d: %w", id, err)
	}
	return user, nil
}

// FindAll returns all persisted users.
func (d *sqlUserDao) FindAll(ctx context.Context) ([]*model.User, error) {
	const query = `SELECT id, name FROM users`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate users: %w", err)
	}
	return users, nil
}

// Save inserts a new user or updates an existing one.
//
// MIGRATION_NOTE: JpaRepository.save performed an upsert based on whether the
// entity's ID was set. This implementation mirrors that behaviour: a zero ID
// triggers an INSERT (and the generated ID is populated), otherwise an UPDATE
// is issued. The exact ID/Name access on model.User must be reconciled with
// the final model definition during manual review.
func (d *sqlUserDao) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, errors.New("repository: user must not be nil")
	}

	// MIGRATION_NOTE: model.User field accessors are assumed here. Replace
	// user.ID / user.Name with the actual exported fields once model.User is
	// finalized.
	if user.ID == 0 {
		const insert = `INSERT INTO users (name) VALUES (?)`
		res, err := d.db.ExecContext(ctx, insert, user.Name)
		if err != nil {
			return nil, fmt.Errorf("repository: insert user: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("repository: read generated user id: %w", err)
		}
		user.ID = int(id)
		return user, nil
	}

	const update = `UPDATE users SET name = ? WHERE id = ?`
	if _, err := d.db.ExecContext(ctx, update, user.Name, user.ID); err != nil {
		return nil, fmt.Errorf("repository: update user %d: %w", user.ID, err)
	}
	return user, nil
}

// DeleteByID removes the user with the given primary key.
func (d *sqlUserDao) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = ?`

	res, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user %d: rows affected: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("repository: delete user %d: %w", id, smartcontacterror.ErrUserNotFound)
	}
	return nil
}

// rowScanner abstracts *sql.Row and *sql.Rows so a single scan helper works
// for both single-row and multi-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser reads a single user row into a model.User.
//
// MIGRATION_NOTE: The scan targets (&user.ID, &user.Name) assume model.User
// exposes ID and Name fields. Reconcile with the final model.User definition.
func scanUser(s rowScanner) (*model.User, error) {
	var user model.User
	if err := s.Scan(&user.ID, &user.Name); err != nil {
		return nil, err
	}
	return &user, nil
}
