// Package smartcontact is the application composition root for the SmartContact
// service. In the original Spring Boot application this responsibility lived in
// SmartContactApplication.java, annotated with @SpringBootApplication, whose sole
// job was to bootstrap the Spring ApplicationContext, run component scanning, and
// start the embedded web server.
//
// MIGRATION_NOTE: Go has no equivalent to Spring's auto-configuration or
// classpath component scanning. Dependency wiring is explicit in Go, so instead
// of magic annotations we expose buildRouter(), which constructs the fully-wired
// chi router. The actual process entry point (func main) lives in
// cmd/server/main.go, which is expected to call buildRouter() to obtain the
// http.Handler and serve it with graceful shutdown.
//
// REQUIRES MANUAL REVIEW:
//   - As real controllers are migrated, replace the placeholder stub handlers
//     below with the concrete handler functions from their respective files and
//     inject any required dependencies (DB, services) through a router builder
//     that accepts them as arguments.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the
// SmartContact application.
//
// It is the Go equivalent of Spring Boot's auto-configured DispatcherServlet plus
// component-scanned @Controller beans: every route the application serves is
// registered here explicitly. cmd/server/main.go calls buildRouter() directly to
// obtain the http.Handler it serves.
//
// MIGRATION_NOTE: The source SmartContactApplication.java defined no routes of
// its own — it merely triggered component scanning of the com.smartContact
// package. Because the controller source files are not part of this migration
// batch, their routes are wired here as clearly-labelled placeholder stubs
// (returning 200 OK) so the router is reachable and testable. Replace each stub
// with the migrated handler as its owning controller file is ported.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	// Cross-cutting middleware. Equivalent to Spring Boot's built-in request
	// logging and the framework's exception-handling safety net.
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe. Required by deployment tooling.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// --- Public routes (owned by not-yet-migrated controllers) ---
	// MIGRATION_NOTE: placeholder stubs — wire real handlers as controllers land.
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

	// --- Authenticated user area (owned by not-yet-migrated controllers) ---
	// MIGRATION_NOTE: Spring Security guarded /user/** for authenticated users.
	// Add authentication middleware here once the security layer is migrated.
	r.Route("/user", func(r chi.Router) {
		r.Get("/index", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: UserDashboard")
		})
	})

	// --- Admin area ---
	// MIGRATION_NOTE: admin-prefixed routes belong to not-yet-migrated
	// controllers; add role-based authorization middleware when available.
	r.Route("/admin", func(r chi.Router) {
		r.Get("/index", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDashboard")
		})
	})

	return r
}

// BuildRouter is the exported entry point cmd/server/main.go calls to obtain
// the fully-wired HTTP handler.
func BuildRouter() http.Handler {
	return buildRouter()
}
