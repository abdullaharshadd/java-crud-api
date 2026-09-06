package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func (ur *UserRepository) GetUser(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := ur.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id=$1", id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pq.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}