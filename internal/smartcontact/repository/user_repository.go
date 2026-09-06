```go
package repository

import (
	"context"
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	"github.com/jmoiron/sqlx"
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
	if err == sql.ErrNoRows {
		return nil, NewUserNotFoundError()
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```