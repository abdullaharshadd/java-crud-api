package handler

import (
	"context"
	"net/http"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := uc.userService.GetUser(context.Background(), id)
	if err != nil {
		log.Err(err).Msg("failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}