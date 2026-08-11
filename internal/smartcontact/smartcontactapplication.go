// Package smartcontact is the application composition root migrated from
// com.smartContact.SmartContactApplication.
//
// MIGRATION_NOTE: The Java source was a Spring Boot @SpringBootApplication
// main class. It relied on convention-over-configuration: component
// scanning wired the @RestController, @Service and @Repository beans, the
// embedded server was started implicitly, and the schema was managed by
// Hibernate ddl-auto.
//
// Go has no IoC container or auto-configuration, so the wiring is explicit:
//   - buildRouter() constructs the dependency graph (repository -> service
//     -> handler) and returns a fully-configured chi router.
//   - The database handle and schema creation are set up here at startup
//     (EnsureUserSchema), replacing Hibernate's ddl-auto.
//   - cmd/server/main.go owns func main(), opens the http.Server and
//     performs graceful shutdown; it calls BuildRouter() from this package.
package smartcontact

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

// DefaultDatabaseURL is the DSN used when SMARTCONTACT_DATABASE_URL is unset.
// MIGRATION_NOTE: In the Java app this came from application.properties;
// here it is read from the environment with a sane local default.
const DefaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/smartcontact?sslmode=disable"

// openDB opens the PostgreSQL connection pool and verifies connectivity.
// The caller owns the returned *sql.DB and is responsible for closing it.
func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("SMARTCONTACT_DATABASE_URL")
	}
	if dsn == "" {
		dsn = DefaultDatabaseURL
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// BuildRouter constructs the full dependency graph and returns a ready-to-serve
// http.Handler. It is the explicit replacement for Spring's component scanning
// and embedded-server bootstrap. cmd/server/main.go calls this directly.
//
// MIGRATION_NOTE: BuildRouter must not fail the whole process on a missing
// database at import time, but a nil handler graph would be worse. If the DB
// cannot be reached the router still serves /healthz and returns 503 for the
// user routes, so the server can start and be diagnosed. The returned handler
// keeps its own *sql.DB alive for the lifetime of the process.
func BuildRouter() http.Handler {
	return buildRouter()
}

// buildRouter is the internal implementation used by BuildRouter and tests.
func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	db, err := openDB()
	if err != nil {
		// Degraded mode: report the failure on every user route rather than
		// panicking, so /healthz and diagnostics remain available.
		r.HandleFunc("/*", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, fmt.Sprintf("database unavailable: %v", err), http.StatusServiceUnavailable)
		})
		return r
	}

	// Schema creation: explicit replacement for Hibernate ddl-auto.
	if err := model.EnsureUserSchema(db); err != nil {
		r.HandleFunc("/*", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, fmt.Sprintf("schema initialization failed: %v", err), http.StatusServiceUnavailable)
		})
		return r
	}

	// Constructor injection: repository -> service -> handler.
	userRepo := repository.NewUserDao(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// The handler owns its route registration (migrated RegisterRoutes),
	// mirroring the annotation-driven mappings from the Java UserController.
	userHandler.RegisterRoutes(r)

	return r
}
