// Package smartcontact is the application entry point for the SmartContact
// service. It corresponds to the original Spring Boot main class
// (com.smartContact.SmartContactApplication).
//
// MIGRATION_NOTE: The Spring Boot bootstrap model (@SpringBootApplication with
// component scanning and an embedded server started via SpringApplication.run)
// has no direct Go equivalent. In idiomatic Go there is no auto-configuration
// or component scanning: dependencies are wired explicitly, and the HTTP server
// is started from cmd/server/main.go. The role of the old main class is split:
//
//   - cmd/server/main.go owns process startup, config loading, DB connection,
//     graceful shutdown, and calls buildRouter() to obtain the HTTP handler.
//   - buildRouter() (below) assembles the chi router, middleware, and routes,
//     replacing Spring's implicit dispatcher + component scan.
//
// As controllers are migrated, wire their real handlers into buildRouter().
// Routes for controllers not yet migrated in this batch use explicit 200-OK
// stub handlers so the router stays reachable and testable.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the
// SmartContact application. It is invoked directly by cmd/server/main.go.
//
// This replaces Spring Boot's implicit DispatcherServlet + component scanning:
// middleware and routes are registered explicitly here. As individual
// controllers are migrated, replace the corresponding stub handlers with the
// real migrated handler functions.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	// Standard middleware chain (replaces Spring's servlet filters):
	// request logging and panic recovery.
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe endpoint.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// MIGRATION_NOTE: The source SmartContactApplication.java is the entry point
	// only and defines no routes of its own. The route table below is assembled
	// from the controllers of this application. Because no controller files were
	// present in this migration batch, the concrete application routes are wired
	// as explicit 200-OK stubs and MUST be replaced with the real migrated
	// handlers as each controller is migrated. Do not remove these stubs before
	// their owning controllers are wired in, or those paths become unreachable.

	// Public / home routes (owned by a not-yet-migrated home controller).
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: Home")
	})
	r.Get("/signup", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: Signup")
	})
	r.Post("/do_register", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: DoRegister")
	})
	r.Get("/signin", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: Signin")
	})

	// Admin-prefixed routes (owned by a not-yet-migrated admin/user controller).
	r.Route("/admin", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDashboard")
		})
	})

	return r
}
