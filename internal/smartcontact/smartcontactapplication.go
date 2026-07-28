package smartcontact

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	apperror "migrated-app/internal/smartcontact/error/restresponseentityexceptionhandling"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

// buildRouter constructs the fully-wired HTTP handler for the smartContact
// application. It is the Go equivalent of Spring Boot's application bootstrap:
// it builds the dependency graph (repository -> service -> controller) and
// registers every route on a chi router with logging, panic-recovery, and
// application error-mapping middleware.
//
// It is intentionally exported-by-convention for use from cmd/server/main.go,
// which owns the *http.Server and graceful-shutdown lifecycle.
func buildRouter() http.Handler {
	// MIGRATION_NOTE: The datasource was configured by Spring at runtime from
	// application.properties; the source entry point held no explicit *sql.DB.
	// cmd/server/main.go should open the *sql.DB and pass it into the wiring.
	// Until that is done this uses a nil handle, which will fail at query time,
	// not at startup. See requires_manual_review.
	var db *sql.DB
	return newRouter(db)
}

// newRouter builds the router with an explicit *sql.DB dependency so callers
// (including tests) can inject a real or mock database. buildRouter is a thin
// wrapper that supplies the process-wide connection.
func newRouter(db *sql.DB) http.Handler {
	userRepo := repository.NewUserDao(db)
	userService := service.NewUserServiceImp(userRepo)

	// The handler.UserService interface uses goerror return types while
	// service.UserServiceImp uses error. We adapt with a thin wrapper.
	userController := handler.NewUserController(&userServiceAdapter{userService}, nil)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Application-specific error mapping (@ControllerAdvice equivalent):
	// translates UserNotFoundError panics into structured 404 responses.
	r.Use(apperror.RecoverMiddleware)

	// Liveness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Routes owned by this migration batch (UserController). RegisterRoutes
	// wires POST /save_user_data, GET /get_user_data, GET /get_user_data/{id},
	// DELETE /delete_user_data/{id}, PUT /update_user_data/{id}, and
	// GET /get_user_name/name/{name}.
	userController.RegisterRoutes(r)

	return r
}

// userServiceAdapter adapts *service.UserServiceImp (which returns error) to
// handler.UserService (which returns goerror = interface{ Error() string }).
// Both error and goerror are structurally identical interfaces, but Go's type
// system requires an explicit adapter when the declared return types differ.
type userServiceAdapter struct {
	impl service.UserService
}

func (a *userServiceAdapter) SaveUser(ctx context.Context, user model.User) error {
	_, err := a.impl.SaveUser(ctx, user)
	return err
}

func (a *userServiceAdapter) FetchUserList(ctx context.Context) ([]model.UserResponse, error) {
	return a.impl.FetchUserList(ctx)
}

func (a *userServiceAdapter) FetchUserByID(ctx context.Context, id int) (model.UserResponse, error) {
	return a.impl.FetchUserByID(ctx, id)
}

func (a *userServiceAdapter) DeleteUser(ctx context.Context, id int) error {
	return a.impl.DeleteUser(ctx, id)
}

func (a *userServiceAdapter) UpdateUser(ctx context.Context, id int, user model.User) error {
	return a.impl.UpdateUser(ctx, id, user)
}

func (a *userServiceAdapter) GetUserByName(ctx context.Context, name string) (model.UserResponse, error) {
	return a.impl.GetUserByName(ctx, name)
}