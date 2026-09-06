package repository

import (
	"context"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUser(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	err := ur.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if err == sqlx.ErrNoRows {
			return nil, UserNotFoundError{}
		}
		log.Err(err).Msg("failed to get user")
		return nil, err
	}
	return &user, nil
}

type UserNotFoundError struct{}

func (e UserNotFoundError) Error() string {
	return "User not found"
}