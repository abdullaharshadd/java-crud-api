package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"migrated-app/internal/smartcontact/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) SaveUser(ctx context.Context, user *model.User) error {
	// Save user logic
	return nil
}

func (ur *UserRepository) FetchUserList(ctx context.Context) ([]*model.User, error) {
	// Fetch user list logic
	return nil, nil
}

func (ur *UserRepository) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	// Fetch user by ID logic
	return nil, ErrUserNotFound
}

func (ur *UserRepository) DeleteUser(ctx context.Context, id int) error {
	// Delete user logic
	return nil
}

func (ur *UserRepository) UpdateUser(ctx context.Context, id int, user *model.User) error {
	// Update user logic
	return nil
}

func (ur *UserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	// Find user by name logic
	return nil, ErrUserNotFound
}

func IsUserNotFoundError(err error) bool {
	return err == ErrUserNotFound
}