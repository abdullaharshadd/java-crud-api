package repository

import (
	"context"
	"migrated-app/internal/smartcontact/model"
	"database/sql"
	_ "github.com/lib/pq"
)

var db *sql.DB

func GetUser(id string) (*model.User, error) {
	// implementation
	return &model.User{}, nil
}

func UpdateUser(user model.User) error {
	// implementation
	return nil
}

func DeleteUser(id string) error {
	// implementation
	return nil
}