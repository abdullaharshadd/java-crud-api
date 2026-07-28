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

	adapted := &userServiceAdapter{impl: userService}
	var svc handler.UserService = adapted
	userController := handler.NewUserController(svc, nil)

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

type userServiceAdapter struct {
	impl service.UserService
}

func (a *userServiceAdapter) SaveUser(ctx context.Context, user model.User) interface{ Error() string } {
	_, err := a.impl.SaveUser(ctx, user)
	if err == nil {
		return nil
	}
	return err
}

func (a *userServiceAdapter) FetchUserList(ctx context.Context) ([]model.UserResponse, interface{ Error() string }) {
	res, err := a.impl.FetchUserList(ctx)
	if err == nil {
		return res, nil
	}
	return res, err
}

func (a *userServiceAdapter) FetchUserByID(ctx context.Context, id int) (model.UserResponse, interface{ Error() string }) {
	res, err := a.impl.FetchUserByID(ctx, id)
	if err == nil {
		return res, nil
	}
	return res, err
}

func (a *userServiceAdapter) DeleteUser(ctx context.Context, id int) interface{ Error() string } {
	err := a.impl.DeleteUser(ctx, id)
	if err == nil {
		return nil
	}
	return err
}

func (a *userServiceAdapter) UpdateUser(ctx context.Context, id int, user model.User) interface{ Error() string } {
	err := a.impl.UpdateUser(ctx, id, user)
	if err == nil {
		return nil
	}
	return err
}

func (a *userServiceAdapter) GetUserByName(ctx context.Context, name string) (model.UserResponse, interface{ Error() string }) {
	res, err := a.impl.GetUserByName(ctx, name)
	if err == nil {
		return res, nil
	}
	return res, err
}