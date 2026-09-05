package handler

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
)

type UserController struct {
	UserRepo repository.UserRepository
}

func (uc *UserController) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	return uc.UserRepo.GetUserByID(ctx, id)
}

func NewUserController(userRepo repository.UserRepository) *UserController {
	return &UserController{UserRepo: userRepo}
}