package repository

import (
	"context"
	"database/sql"
	"errors"
	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the methods for interacting with the user table.
type UserRepository interface {
	FindByName(context.Context, string) (*model.User, error)
}

// newUserRepository creates a new UserRepository instance.
func newUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// userRepository is the implementation of UserRepository.
type userRepository struct {
	db *sql.DB
}

// FindByName retrieves a user by their name.
func (ur *userRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	query := "SELECT id, name, email, password, role, about FROM users WHERE name = $1"
	row := ur.db.QueryRowContext(ctx, query, name)
	var user model.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.About); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewUserNotFoundError("User not found", nil)
		}
		return nil, NewUserNotFoundError("Failed to retrieve user", err)
	}
	return &user, nil
}
