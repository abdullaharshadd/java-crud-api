package repository

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the full persistence operations the service layer needs.
// This extends the interface defined in userdao.go to include FindAll.

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	Save(ctx context.Context, user *model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	DeleteByID(ctx context.Context, id int64) error
}