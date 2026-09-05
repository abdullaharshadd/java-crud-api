package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	_ "github.com/lib/pq"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	// Implementation of GetUserByID
	return nil, nil
}