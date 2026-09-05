package handler

import (
	"migrated-app/internal/smartcontact/service"
	"migrated-app/internal/smartcontact/repository"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService service.UserService
}

func NewUserController(userRepo repository.UserRepository) *UserController {
	return &UserController{
		UserService: service.NewUserService(userRepo),
	}
}

func (uc *UserController) CreateUser(c *gin.Context) {
	// Create user logic
}

func (uc *UserController) GetUser(c *gin.Context) {
	// Get user logic
}