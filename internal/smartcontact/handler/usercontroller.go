```go
package handler

import (
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// UserController handles user-related operations.
type UserController struct {
	userService service.UserService
}

// NewUserController creates a new instance of UserController.
func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// CreateUser handles the creation of a new user.
func (uc *UserController) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Error().Err(err).Msg("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	if err := user.Validate(); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newUser, err := uc.userService.CreateUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, newUser)
}

// GetUser retrieves a user by ID.
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userService.GetUser(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
```