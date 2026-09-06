package handler

import (
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
)

type UserController struct {
	userRepo repository.UserRepository
}

func (uc *UserController) GetUser(ctx context.Context, id int) (*model.User, error) {
	return uc.userRepo.GetUser(ctx, id)
}