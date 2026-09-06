package repository

import (
	"errors"
	"migrated-app/internal/smartcontact/model"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var (
	UserNotFoundError = errors.New("user not found")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUserByID(id int) (*model.User, error) {
	var user model.User
	query := "SELECT id, username, email FROM users WHERE id=$1;"
	err := ur.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, UserNotFoundError
	}
	if err != nil {
		log.Err(err).Msg("failed to get user by ID")
		return nil, err
	}
	return &user, nil
}

func IsUserNotFoundError(err error) bool {
	return err == UserNotFoundError
}