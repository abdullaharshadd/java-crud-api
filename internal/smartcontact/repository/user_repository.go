package repository

import (
	"context"
	"database/sql"

	apperror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	GetUser(context.Context, int) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) GetUser(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	query := "SELECT id, name, email FROM users WHERE id=$1;"
	err := ur.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err == sql.ErrNoRows {
		return nil, apperror.NewUserNotFoundError()
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsUserNotFoundError returns true if err is a UserNotFoundError.
func IsUserNotFoundError(err error) bool {
	return apperror.IsUserNotFoundError(err)
}