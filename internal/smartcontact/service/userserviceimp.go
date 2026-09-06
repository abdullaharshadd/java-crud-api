package service

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
)

type UserServiceImp struct {
	userRepo repository.UserRepository
}

func NewUserServiceImp(userRepo repository.UserRepository) *UserServiceImp {
	return &UserServiceImp{userRepo: userRepo}
}

func (usi *UserServiceImp) GetUser(ctx context.Context, id int) (*repository.User, error) {
	return usi.userRepo.GetUser(ctx, id)
}