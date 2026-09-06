package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
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
	// implementation here
	return nil, nil
}