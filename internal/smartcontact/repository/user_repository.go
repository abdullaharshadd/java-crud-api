package repository

import (
	"context"
	"database/sql"
	"errors"
	"migrated-app/internal/smartcontact/model"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func (ur *UserRepository) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	user := &model.User{}
	query := "SELECT * FROM users WHERE id = $1"
	err := ur.db.GetContext(ctx, user, query, id)
	if err == sql.ErrNoRows {
		return nil, UserNotFoundError
	} else if err != nil {
		return nil, err
	}
	return user, nil
}