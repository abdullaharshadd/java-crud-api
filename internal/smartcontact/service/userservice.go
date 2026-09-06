package service

import (
	"context"

	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserService defines the interface for user business logic.
type UserService interface {
	GetUser(ctx context.Context, id int) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (us *userService) GetUser(ctx context.Context, id int) (*model.User, error) {
	return us.userRepo.GetUser(ctx, id)
}