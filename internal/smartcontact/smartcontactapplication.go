// Package smartcontact is the composition root for the smartContact
// application. It wires together the persistence, service, and HTTP handler
// layers and exposes a fully-configured http.Handler via buildRouter.
//
// MIGRATION_NOTE: The Java source (SmartContactApplication) was the Spring Boot
// entry point annotated with @SpringBootApplication. Spring performed component
// scanning, auto-configuration, dependency injection, and embedded-server
// bootstrap implicitly. Go has none of that magic, so this file makes the
// wiring explicit:
//   - Component scanning / @Autowired      -> manual constructor injection.
//   - @EnableAutoConfiguration + Tomcat    -> chi.Router built here and served
//                                             by cmd/server/main.go.
//   - @ControllerAdvice exception mapping  -> RecoverMiddleware + WriteError.
//
// MIGRATION_NOTE: The original SpringApplication.run(...) call lived in a
// static main(). Per the target project layout the actual process entry point
// is cmd/server/main.go, which calls BuildRouter() to obtain the wired handler.
// This file therefore does NOT declare package main or func main.
package smartcontact

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	apperror "migrated-app/internal/smartcontact/error/restresponseentityexceptionhandling"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

// BuildRouter constructs the fully-wired HTTP handler for the smartContact
// application using the provided *sql.DB. It is exported so that
// cmd/server/main.go can obtain the wired handler after opening the database
// connection from configuration.
func BuildRouter(db *sql.DB) http.Handler {
	return newRouter(db)
}

// buildRouter constructs the fully-wired HTTP handler for the smartContact
// application. It is the Go equivalent of Spring Boot's application bootstrap:
// it builds the dependency graph (repository -> service -> controller) and
// registers every route on a chi router with logging, panic-recovery, and
// application error-mapping middleware.
//
// It is intentionally kept for internal use and tests; external callers use
// BuildRouter.
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
	userController := handler.NewUserController(userService, nil)

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
