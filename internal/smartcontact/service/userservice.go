package service

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (us *UserService) GetUser(ctx context.Context, id int) (*repository.User, error) {
	return us.userRepo.GetUser(ctx, id)
}