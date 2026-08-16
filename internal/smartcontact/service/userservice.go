package service

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserService defines the business-logic operations for user management.
type UserService interface {
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)
	FetchUserList(ctx context.Context) ([]*model.User, error)
	FetchUserByID(ctx context.Context, id int64) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateUser(ctx context.Context, id int64, user *model.User) error
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}