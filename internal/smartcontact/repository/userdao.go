package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"migrated-app/internal/smartcontact/model"
)

// UserDao is the concrete PostgreSQL-backed repository for User records.
type UserDao struct {
	db *sql.DB
}

// NewUserDao constructs a UserDao backed by the given *sql.DB.
func NewUserDao(db *sql.DB) *UserDao {
	return &UserDao{db: db}
}

// Save inserts or updates a user. When user.ID == 0 an INSERT is performed and
// the generated id is returned. When user.ID != 0 an UPDATE is performed.
func (r *UserDao) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if user.ID == 0 {
		const q = `INSERT INTO users (name, email, password, role, about)
		           VALUES ($1, $2, $3, $4, $5)
		           RETURNING id, name, email, password, role, about`
		row := r.db.QueryRowContext(ctx, q,
			user.Name, user.Email, user.Password, user.Role, user.About)
		out := &model.User{}
		if err := row.Scan(&out.ID, &out.Name, &out.Email, &out.Password, &out.Role, &out.About); err != nil {
			return nil, fmt.Errorf("save (insert) user: %w", err)
		}
		return out, nil
	}
	const q = `UPDATE users SET name=$1, email=$2, password=$3, role=$4, about=$5
	           WHERE id=$6
	           RETURNING id, name, email, password, role, about`
	row := r.db.QueryRowContext(ctx, q,
		user.Name, user.Email, user.Password, user.Role, user.About, user.ID)
	out := &model.User{}
	if err := row.Scan(&out.ID, &out.Name, &out.Email, &out.Password, &out.Role, &out.About); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("save (update) user %d: not found", user.ID)
		}
		return nil, fmt.Errorf("save (update) user %d: %w", user.ID, err)
	}
	return out, nil
}

// FindAll returns all users in the store.
func (r *UserDao) FindAll(ctx context.Context) ([]model.User, error) {
	const q = `SELECT id, name, email, password, role, about FROM users ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About); err != nil {
			return nil, fmt.Errorf("find all users scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find all users rows: %w", err)
	}
	return users, nil
}

// FindByID returns the user with the given id, or nil if none exists.
func (r *UserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	const q = `SELECT id, name, email, password, role, about FROM users WHERE id=$1`
	row := r.db.QueryRowContext(ctx, q, id)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id %d: %w", id, err)
	}
	return u, nil
}

// DeleteByID removes the user with the given id.
func (r *UserDao) DeleteByID(ctx context.Context, id int) error {
	const q = `DELETE FROM users WHERE id=$1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// FindByName returns the first user whose name matches, or nil if none exists.
func (r *UserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	const q = `SELECT id, name, email, password, role, about FROM users WHERE name=$1 LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, name)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by name %q: %w", name, err)
	}
	return u, nil
}