package handler

import (
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/service"
	"net/http"
	"strconv"
	"context"
)

func NewUserController(svc service.UserService) *UserController {
	return &UserController{userService: svc}
}

type UserController struct {
	userService service.UserService
}

func (uc *UserController) RegisterRoutes(r *gin.Engine) {
	r.GET("/users/:id", uc.GetUserHandler)
}

func (uc *UserController) GetUserHandler(c *gin.Context) {
	ctx := context.TODO()
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		error.HandleError(c.Writer, err)
		return
	}
	user, err := uc.userService.GetUser(ctx, userID)
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, user)
}