package smartcontact

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	errormodel "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/service"
)

// NewUserController creates a new UserController.
func NewUserController(svc service.UserService) *UserController {
	return &UserController{userService: svc}
}

// UserController handles HTTP requests for users.
type UserController struct {
	userService service.UserService
}

// RegisterRoutes registers the routes for the UserController on a gin.Engine.
func (uc *UserController) RegisterRoutes(r *gin.Engine) {
	r.GET("/users/:id", uc.GetUserHandler)
}

// GetUserHandler handles GET /users/:id.
func (uc *UserController) GetUserHandler(c *gin.Context) {
	ctx := context.TODO()
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		apperror.HandleError(c.Writer, err)
		return
	}
	user, err := uc.userService.GetUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}