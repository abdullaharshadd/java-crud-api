package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/error"
)

type UserRepository interface {
	GetUser(context.Context, int) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) GetUser(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	query := "SELECT id, name, email FROM users WHERE id=$1;"
	err := ur.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err == sql.ErrNoRows {
		return nil, error.NewUserNotFoundError()
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func IsUserNotFoundError(err error) bool {
	_, ok := err.(*error.UserNotFoundError)
	return ok
}