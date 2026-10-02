package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"migrated-app/internal/model"
)

// UserRepository provides persistence operations for model.User, replacing the
// Spring Data JpaRepository<User, Integer> interface with hand-written SQL.
//
// MIGRATION_NOTE: Spring Data derived/inherited methods (save, findById,
// findAll, deleteById, delete, count, existsById, findByName) are implemented
// explicitly here. The source relied on an auto-generated proxy; Go has no such
// mechanism, so each operation is written by hand against sqlx.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a UserRepository backed by the given sqlx.DB and
// ensures the underlying schema exists.
//
// MIGRATION_NOTE: Hibernate ddl-auto=update created/updated the USER table at
// boot. That responsibility moves here: the CREATE TABLE IF NOT EXISTS below
// derives its columns directly from model.User's real fields.
func NewUserRepository(ctx context.Context, db *sqlx.DB) (*UserRepository, error) {
	r := &UserRepository{db: db}
	if err := r.ensureSchema(ctx); err != nil {
		return nil, fmt.Errorf("ensure user schema: %w", err)
	}
	return r, nil
}

// ensureSchema creates the USER table if it does not already exist. Columns
// mirror model.User exactly: id (int, auto-increment PK) plus the nullable
// string fields name, email, password, role, about.
func (r *UserRepository) ensureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS USER (
    id       INT AUTO_INCREMENT PRIMARY KEY,
    name     VARCHAR(255) NULL,
    email    VARCHAR(255) NULL,
    password VARCHAR(255) NULL,
    role     VARCHAR(255) NULL,
    about    VARCHAR(255) NULL
)`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("create USER table: %w", err)
	}
	return nil
}

// Save persists a new User, populating its generated ID from LastInsertId.
//
// MIGRATION_NOTE: JpaRepository.save() is upsert-like; the insert-only path is
// Save here, while the upsert path lives in Merge (see below).
func (r *UserRepository) Save(ctx context.Context, u *model.User) (*model.User, error) {
	const q = `INSERT INTO USER (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q,
		u.GetName(), u.GetEmail(), u.GetPassword(), u.GetRole(), u.GetAbout())
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("save user last insert id: %w", err)
	}
	u.SetID(int(id))
	return u, nil
}

// Merge persists or updates a User, mirroring JpaRepository.save()'s upsert
// semantics.
//
// MIGRATION_NOTE: ON DUPLICATE KEY UPDATE was explicitly rejected in favor of a
// PK-exists check inside a transaction, then UPDATE or INSERT, to faithfully
// match JPA's merge behavior (which keys off the entity's identifier).
func (r *UserRepository) Merge(ctx context.Context, u *model.User) (*model.User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("merge user begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if u.GetID() != 0 {
		var exists bool
		if err := tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM USER WHERE id = ?)`, u.GetID()); err != nil {
			return nil, fmt.Errorf("merge user exists check: %w", err)
		}
		if exists {
			const upd = `UPDATE USER SET name = ?, email = ?, password = ?, role = ?, about = ? WHERE id = ?`
			if _, err := tx.ExecContext(ctx, upd,
				u.GetName(), u.GetEmail(), u.GetPassword(), u.GetRole(), u.GetAbout(), u.GetID()); err != nil {
				return nil, fmt.Errorf("merge user update: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("merge user commit: %w", err)
			}
			return u, nil
		}
	}

	const ins = `INSERT INTO USER (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, ins,
		u.GetName(), u.GetEmail(), u.GetPassword(), u.GetRole(), u.GetAbout())
	if err != nil {
		return nil, fmt.Errorf("merge user insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("merge user last insert id: %w", err)
	}
	u.SetID(int(id))
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("merge user commit: %w", err)
	}
	return u, nil
}

// FindByID retrieves a User by its integer id. The boolean return is false when
// no row matches (sql.ErrNoRows), mirroring JpaRepository.findById returning an
// empty Optional.
func (r *UserRepository) FindByID(ctx context.Context, id int) (*model.User, bool, error) {
	var u model.User
	const q = `SELECT id, name, email, password, role, about FROM USER WHERE id = ?`
	if err := r.db.GetContext(ctx, &u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("find user by id: %w", err)
	}
	return &u, true, nil
}

// FindAll retrieves every User. The returned slice is always non-nil, even when
// empty, mirroring JpaRepository.findAll returning an empty List.
func (r *UserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	users := make([]model.User, 0)
	const q = `SELECT id, name, email, password, role, about FROM USER`
	if err := r.db.SelectContext(ctx, &users, q); err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	return users, nil
}

// FindByName retrieves a single User matching the given name.
//
// MIGRATION_NOTE: Signed-off divergence from JPA. The derived query
// findByName(String) would throw NonUniqueResultException on multiple matches;
// here LIMIT 1 is used instead. apperr.ErrUserNotFound is returned when no row
// matches.
func (r *UserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	const q = `SELECT id, name, email, password, role, about FROM USER WHERE name = ? LIMIT 1`
	if err := r.db.GetContext(ctx, &u, q, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by name: %w", err)
	}
	return &u, nil
}

// DeleteByID deletes the User with the given integer id.
//
// MIGRATION_NOTE: When zero rows are affected a non-sentinel error is returned
// (surfaced as HTTP 500 upstream), preserving the source's behavior.
func (r *UserRepository) DeleteByID(ctx context.Context, id int) error {
	const q = `DELETE FROM USER WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete user by id: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user by id rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("delete user by id: no rows affected for id %d", id)
	}
	return nil
}

// Delete removes the given User entity by its id, mirroring
// JpaRepository.delete(entity).
func (r *UserRepository) Delete(ctx context.Context, u *model.User) error {
	return r.DeleteByID(ctx, u.GetID())
}

// Count returns the total number of User entities, mirroring
// JpaRepository.count().
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	const q = `SELECT COUNT(*) FROM USER`
	if err := r.db.GetContext(ctx, &n, q); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

// ExistsByID reports whether a User with the given id exists, mirroring
// JpaRepository.existsById().
func (r *UserRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	var exists bool
	const q = `SELECT EXISTS(SELECT 1 FROM USER WHERE id = ?)`
	if err := r.db.GetContext(ctx, &exists, q, id); err != nil {
		return false, fmt.Errorf("exists user by id: %w", err)
	}
	return exists, nil
}