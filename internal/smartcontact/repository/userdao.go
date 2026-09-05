package repository

import (
	"database/sql"
	"migrated-app/internal/smartcontact/model"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	// implementation
	return nil, nil
}