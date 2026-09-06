package service

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
)

type UserServiceImp struct {
	userRepo repository.UserRepository
}

func NewUserServiceImp(userRepo repository.UserRepository) *UserServiceImp {
	return &UserServiceImp{userRepo: userRepo}
}

func (usi *UserServiceImp) GetUser(ctx context.Context, id int) (*model.User, error) {
	return usi.userRepo.GetUser(ctx, id)
}