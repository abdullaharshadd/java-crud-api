// Package smartcontact is the composition root for the SmartContact
// application. It wires together the repository, service, and handler layers
// and exposes a fully-configured HTTP router.
//
// MIGRATION_NOTE: The Java source was the @SpringBootApplication main class.
// Spring Boot's SpringApplication.run() performed classpath scanning,
// auto-configuration, bean instantiation and dependency injection, and
// embedded-server startup all implicitly. Go has no equivalent runtime magic,
// so all of that is replaced with explicit, compile-time dependency injection
// here in buildRouter(): we construct the repository, inject it into the
// service, inject the service into the controller, and register the
// controller's routes on a chi router.
//
// MIGRATION_NOTE: The Java main() method lived in this class. In idiomatic Go
// the process entry point (package main / func main) lives in
// cmd/server/main.go, which calls buildRouter() to obtain the http.Handler and
// serves it (typically with graceful shutdown). This file therefore exposes
// buildRouter() rather than declaring its own main().
//
// MIGRATION_NOTE: The database connection (*sql.DB) is a required dependency of
// the repository layer. Spring auto-configured the DataSource from
// application.properties. Here it must be provided explicitly; buildRouter
// wires a repository backed by whatever *sql.DB is supplied via NewRepository.
// The actual DB open/config belongs in cmd/server/main.go and is passed in.
package smartcontact

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/smartContact/internal/smartcontact/service"
)

// buildRouter constructs the fully-wired HTTP handler for the SmartContact
// application. It performs explicit dependency injection across the
// repository, service, and handler layers, then registers all routes on a chi
// router configured with request logging and panic recovery.
//
// The database handle is read from the package-level DB variable, which the
// process entry point (cmd/server/main.go) is expected to set before calling
// buildRouter. When DB is nil the router is still constructed so that
// infrastructure routes such as /healthz remain reachable.
func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness probe. Kept independent of any downstream dependency so it can
	// report process health even when the database is unavailable.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Explicit dependency injection: repository <- DB, service <- repository,
	// controller <- service. This replaces Spring's @ComponentScan/@Autowired.
	if DB != nil {
		repo := repository.NewUserRepository(DB)
		svc := service.NewUserService(repo)
		ctrl := handler.NewUserController(svc)
		ctrl.RegisterRoutes(r)
	}

	return r
}

// DB is the shared database handle used to construct the repository layer.
//
// MIGRATION_NOTE: In Spring the DataSource was auto-configured and injected.
// Here the process entry point (cmd/server/main.go) must open the *sql.DB
// (PostgreSQL) and assign it to this variable before invoking buildRouter.
// This keeps buildRouter free of connection-string and driver concerns while
// still allowing the fully-wired router to be assembled at startup.
var DB *sql.DB

// SetDB assigns the shared database handle used by buildRouter to wire the
// repository layer. It exists so the entry point can inject the connection
// without directly mutating package state, and returns an error if db is nil.
func SetDB(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("smartcontact: database handle must not be nil")
	}
	DB = db
	return nil
}
