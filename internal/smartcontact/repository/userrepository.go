package repository

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the full persistence operations the service layer needs.
// This extends the interface defined in userdao.go to include FindAll.
type UserRepository interface {
	Save(ctx context.Context, u *model.User) (*model.User, error)
	FindAll(ctx context.Context) ([]*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, bool, error)
	FindByName(ctx context.Context, name string) (*model.User, error)
	Delete(ctx context.Context, id int) error
}