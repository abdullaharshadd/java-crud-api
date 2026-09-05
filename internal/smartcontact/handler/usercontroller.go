package handler

import (
	"context"
	"migrated-app/internal/smartcontact/repository"
)

type UserController struct {
	UserRepo repository.UserRepository
}

func (uc *UserController) GetUserByID(ctx context.Context, id int) (*struct{}, error) {
	return uc.UserRepo.GetUserByID(ctx, id)
}