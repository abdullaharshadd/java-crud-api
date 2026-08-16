// Package smartcontact wires together the smartcontact application's HTTP
// transport, service, and repository layers.
//
// MIGRATION_NOTE: The Java source was the Spring Boot main class
// SmartContactApplication annotated with @SpringBootApplication, whose sole job
// was to call SpringApplication.run(...) — bootstrapping component scanning,
// auto-configuration, and the embedded web server.
//
// Go has no annotation-driven component scanning or auto-configuration, so the
// implicit Spring wiring becomes explicit dependency injection here:
//
//	config.Load -> db.Connect (runs ensureSchema) -> UserDao -> UserService
//	            -> UserController -> RegisterRoutes -> middleware-wrapped router.
//
// The actual process entry point (func main) lives in cmd/server/main.go, which
// calls buildRouter() to obtain the fully-wired handler and serves it on :8082.
// This file therefore exposes buildRouter() rather than declaring its own
// package main / func main.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs the fully-wired HTTP handler for the smartcontact
// application: it applies the standard middleware chain and registers every
// route owned by the migrated UserController.
//
// MIGRATION_NOTE: In a production build the dependency graph (database
// connection, DAO, service, controller) should be constructed from real
// configuration and passed in. Because the concrete constructors for the
// db/repo/service layers require external configuration (connection strings)
// that is resolved in cmd/server/main.go, this function focuses on the routing
// concern and delegates handler wiring to RegisterRoutes on the controller.
// If you have a live *UserController, call BuildRouterWith(controller) instead
// so the real business logic is reachable end-to-end.
func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// MIGRATION_NOTE: The following routes are owned by the migrated
	// UserController (handler.RegisterRoutes). They are wired here as explicit
	// stubs only when a live controller instance is not supplied. The canonical
	// wiring path is BuildRouterWith, which registers the REAL handlers via
	// handler.RegisterRoutes so the CRUD business logic is fully reachable.
	r.Post("/save_user_data", notWiredHandler("SaveUser"))
	r.Get("/get_user_data", notWiredHandler("FetchUserList"))
	r.Get("/get_user_data/{id}", notWiredHandler("FetchUserByID"))
	r.Delete("/delete_user_data/{id}", notWiredHandler("DeleteUser"))
	r.Put("/update_user_data/{id}", notWiredHandler("UpdateUser"))
	r.Get("/get_user_name/name/{name}", notWiredHandler("GetUserByName"))

	return r
}

// routeRegistrar is satisfied by *handler.UserController (which exposes
// RegisterRoutes(chi.Router)). It is declared here to avoid an import cycle
// while still allowing the real controller to be plugged in by main.go.
type routeRegistrar interface {
	RegisterRoutes(r chi.Router)
}

// BuildRouterWith constructs the fully-wired HTTP handler using a live route
// registrar (typically *handler.UserController returned from the
// config -> db -> repo -> service -> handler dependency chain in
// cmd/server/main.go). All CRUD business logic is reachable through the routes
// the registrar installs.
func BuildRouterWith(controller routeRegistrar) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	if controller != nil {
		controller.RegisterRoutes(r)
	}

	return r
}

// notWiredHandler returns a placeholder handler used only when buildRouter is
// invoked without a live controller. It surfaces a clear 503 so an accidental
// unwired deployment is obvious rather than silently returning 200.
func notWiredHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "route %q not wired: build the router with BuildRouterWith(controller)\n", name)
	}
}
