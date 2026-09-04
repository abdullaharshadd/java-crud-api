package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"migrated-app/internal/smartcontact/service"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
)

// UserController handles user-related HTTP requests.
type UserController struct {
	UserService service.UserService
}

// NewUserController creates a new instance of UserController.
func NewUserController(us service.UserService) *UserController {
	return &UserController{
		UserService: us,
	}
}

// SaveUserHandler saves a user to the database.
func (uc *UserController) SaveUserHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	if _, err := uc.UserService.SaveUser(ctx, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully!"})
}

// FetchUserListHandler fetches a list of all users.
func (uc *UserController) FetchUserListHandler(c *gin.Context) {
	ctx := context.Background()
	users, err := uc.UserService.FetchUserList(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// FetchUserByIDHandler fetches a user by ID.
func (uc *UserController) FetchUserByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx := context.Background()
	user, err := uc.UserService.FetchUserByID(ctx, id)
	if repository.IsUserNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUserHandler deletes a user by ID.
func (uc *UserController) DeleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx := context.Background()
	if err := uc.UserService.DeleteUser(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User data deleted successfully!"})
}

// UpdateUserHandler updates a user's information.
func (uc *UserController) UpdateUserHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx := context.Background()
	if err := uc.UserService.UpdateUser(ctx, id, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// FindUserByNameHandler finds a user by their name.
func (uc *UserController) FindUserByNameHandler(c *gin.Context) {
	name := c.Param("name")
	ctx := context.Background()
	user, err := uc.UserService.FindByName(ctx, name)
	if repository.IsUserNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// RegisterRoutes registers all necessary routes for the UserController.
func RegisterRoutes(r *gin.Engine, uc *UserController) {
	r.POST("/save_user_data", uc.SaveUserHandler)
	r.GET("/get_user_data", uc.FetchUserListHandler)
	r.GET("/get_user_data/:id", uc.FetchUserByIDHandler)
	r.DELETE("/delete_user_data/:id", uc.DeleteUserHandler)
	r.PUT("/update_user_data/:id", uc.UpdateUserHandler)
	r.GET("/get_user_name/:name", uc.FindUserByNameHandler)
}