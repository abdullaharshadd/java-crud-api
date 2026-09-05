package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	_ "github.com/lib/pq"
)

type userDao struct {
	db *sql.DB
}

func (ud *userDao) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	// Implementation of GetUserByID
	return nil, nil
}

func NewUserDAO(db *sql.DB) userDao {
	return userDao{db: db}
}