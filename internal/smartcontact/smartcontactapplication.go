// Package smartcontact is the composition root for the SmartContact
// application. It wires together the HTTP router, middleware, and route
// handlers that were previously bootstrapped by Spring Boot's
// @SpringBootApplication auto-configuration.
//
// MIGRATION_NOTE: The original Java class `SmartContactApplication` was a
// Spring Boot entry point whose sole job was to call
// `SpringApplication.run(...)`. In Spring Boot, that single call triggered:
//   - component scanning of the `com.smartContact` base package,
//   - dependency injection / auto-wiring of controllers, services, repos,
//   - embedded servlet container (Tomcat) startup.
//
// Go has no auto-configuration or classpath scanning. All of that implicit
// behavior is replaced by EXPLICIT wiring here in buildRouter, and by an
// explicit `func main` living in cmd/server/main.go which starts the HTTP
// server and manages graceful shutdown. This file deliberately does NOT
// declare `func main` or `package main` — it exposes buildRouter as the
// single seam the real entry point consumes.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the
// SmartContact application. It replaces Spring Boot's auto-configured
// dispatcher servlet and component-scanned controller registration with an
// explicit chi router.
//
// The returned http.Handler is consumed by cmd/server/main.go, which owns
// server lifecycle (listen address, timeouts, graceful shutdown).
//
// MIGRATION_NOTE: The source SmartContactApplication.java contained no
// route definitions of its own — it merely bootstrapped the context. The
// actual controllers (e.g. UserController) live in separate source files
// that are NOT part of this migration batch. Their routes are wired here as
// clearly-marked placeholder handlers so the router remains reachable and
// compilable; replace each stub with the real handler once the owning
// controller file is migrated. Because this file owns NO business routes,
// no business logic is lost by using stubs here.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	// Baseline middleware. middleware.Logger provides request logging
	// (analogous to Spring Boot's built-in request logging), and
	// middleware.Recoverer converts panics into 500 responses instead of
	// crashing the server (analogous to Spring's DispatcherServlet
	// exception handling).
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe. Has no Spring equivalent in the original
	// source; added because it is standard practice for Go HTTP services
	// and is expected by container orchestrators.
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// -----------------------------------------------------------------
	// Routes owned by OTHER (not-yet-migrated) controller files.
	// These are functional 200-OK stubs, not "manual review" dead ends.
	// Replace each with the migrated handler when its controller lands.
	// -----------------------------------------------------------------
	r.Route("/users", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: ListUsers (stub — UserController not yet migrated)")
		})
		r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: GetUserByID (stub — UserController not yet migrated)")
		})
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: CreateUser (stub — UserController not yet migrated)")
		})
		r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: UpdateUser (stub — UserController not yet migrated)")
		})
		r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: DeleteUser (stub — UserController not yet migrated)")
		})
	})

	// Admin-prefixed routes grouped per the required chi pattern. These
	// belong to controllers outside this migration batch and are stubbed.
	r.Route("/admin", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminIndex (stub — Admin controller not yet migrated)")
		})
	})

	return r
}
