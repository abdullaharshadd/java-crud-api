package service

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (us *UserService) GetUser(ctx context.Context, id int) (*model.User, error) {
	return us.userRepo.GetUser(ctx, id)
}