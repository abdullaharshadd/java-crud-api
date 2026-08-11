package repository

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the persistence operations the service layer needs
// for User entities.
type UserRepository interface {
	// Save inserts a new user (when ID is zero) or updates an existing one.
	Save(ctx context.Context, u *model.User) (*model.User, error)
	// FindByID looks up a user by primary key. Returns (nil, false, nil) when no row exists.
	FindByID(ctx context.Context, id int) (*model.User, bool, error)
	// FindAll retrieves all users.
	FindAll(ctx context.Context) ([]*model.User, error)
	// FindByName looks up a user by name. Returns (nil, nil) when no row matches.
	FindByName(ctx context.Context, name string) (*model.User, error)
	// Delete removes the user with the given ID. Returns ErrUserNotFound when no row exists.
	Delete(ctx context.Context, id int) error
}