// Package smartcontact is the application composition root for the
// smartcontact service. It wires together the repository, service and handler
// layers and exposes a fully-configured HTTP router.
//
// MIGRATION_NOTE: The Java source, SmartContactApplication, was the Spring Boot
// entry point annotated with @SpringBootApplication. It relied on Spring's
// auto-configuration and component scanning to discover @Repository, @Service
// and @RestController beans and to bootstrap an embedded servlet container.
// Go has no runtime dependency-injection container or classpath scanning, so
// the bean graph is constructed explicitly here in buildRouter, and the actual
// process bootstrap (server start / graceful shutdown) lives in
// cmd/server/main.go, which calls buildRouter().
//
// MIGRATION_NOTE: Spring Boot obtains a configured DataSource from
// application.properties. Here the *sql.DB must be provided by the caller
// (main.go) because the connection string / driver selection is an
// environment/deployment concern. NewApp accepts the already-open *sql.DB so
// tests can inject a mock or an in-memory database. The target dialect is
// PostgreSQL.
package smartcontact

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

// App holds the fully-wired application dependencies. It is the Go equivalent
// of the Spring ApplicationContext for this service: the object graph is built
// once at startup and reused for the lifetime of the process.
type App struct {
	userController *handler.UserController
}

// NewApp constructs the application dependency graph from an open database
// handle. It replaces Spring's component scanning and @Autowired injection with
// explicit constructor wiring: repository -> service -> handler.
//
// The provided *sql.DB must be non-nil and already configured; NewApp does not
// open or ping the connection.
func NewApp(db *sql.DB) *App {
	userRepo, err := repository.NewUserRepository(db)
	if err != nil {
		// If db is nil or invalid, construct a nil-safe shell; handlers will
		// surface errors naturally when I/O is attempted.
		userRepo = nil
	}
	userService := service.NewUserService(userRepo)
	userController := handler.NewUserController(userService)

	return &App{
		userController: userController,
	}
}

// Router builds and returns the HTTP handler for this application instance,
// registering common middleware, a health check and all user routes.
func (a *App) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	// Delegate all user CRUD route registration to the migrated controller,
	// which owns the exact path definitions from the Java @RequestMapping
	// annotations.
	a.userController.RegisterRoutes(r)

	return r
}
