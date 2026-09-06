package service

import (
	"context"

	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserServiceImp is a concrete implementation of UserService.
type UserServiceImp struct {
	userRepo repository.UserRepository
}

// NewUserServiceImp creates a new UserServiceImp.
func NewUserServiceImp(userRepo repository.UserRepository) UserService {
	return &UserServiceImp{userRepo: userRepo}
}

func (usi *UserServiceImp) GetUser(ctx context.Context, id int) (*model.User, error) {
	return usi.userRepo.GetUser(ctx, id)
}