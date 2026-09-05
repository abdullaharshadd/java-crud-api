package repository

import (
	"context"
	"migrated-app/internal/smartcontact/model"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*model.User, error)
}

type userRepository struct{}

func (ur *userRepository) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	// Implementation of GetUserByID
	return nil, nil
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}