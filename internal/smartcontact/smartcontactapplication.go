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

func buildRouter() http.Handler {
	var db *sql.DB
	return newRouter(db)
}

func newRouter(db *sql.DB) http.Handler {
	userRepo := repository.NewUserDao(db)
	userService := service.NewUserServiceImp(userRepo)

	userController := handler.NewUserController(&userServiceAdapter{userService}, nil)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(apperror.RecoverMiddleware)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	userController.RegisterRoutes(r)

	return r
}

// userServiceAdapter adapts service.UserService (which returns error) to
// handler.UserService (which uses goerror = interface{ Error() string }).
// Because goerror is a type alias for the error interface, both are structurally
// identical, but Go requires the declared return types to match exactly when
// satisfying an interface. The adapter bridges this gap.
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