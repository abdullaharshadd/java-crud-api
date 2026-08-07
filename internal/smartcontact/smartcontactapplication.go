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

	// Delegate all user CRUD route registration to the migrated controller.
	// RegisterRoutes expects *http.ServeMux, so create one and mount it on
	// the chi router to bridge the two routing systems.
	mux := http.NewServeMux()
	a.userController.RegisterRoutes(mux)
	r.Mount("/", mux)

	return r
}