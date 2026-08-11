// Package smartcontact is the application entry point for the Smart Contact
// service. It is the Go equivalent of Spring Boot's @SpringBootApplication main
// class: it wires together the HTTP router, middleware, and (eventually) the
// handler/service/repository dependencies.
//
// MIGRATION_NOTE: The original Java class merely delegated to
// SpringApplication.run(...), which performed component scanning,
// auto-configuration, and embedded-server startup by convention. Go has no
// equivalent "magic" bootstrap: dependency wiring is explicit. The embedded
// server lifecycle (listen + graceful shutdown) lives in cmd/server/main.go,
// which calls buildRouter() from this file to obtain the fully-configured
// http.Handler.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the Smart
// Contact application. cmd/server/main.go calls this directly to obtain the
// http.Handler it serves.
//
// MIGRATION_NOTE: The Smart Contact source class contained no route handlers of
// its own (it was a pure Spring Boot bootstrap class). The routes below are
// placeholder stubs standing in for controllers that belong to OTHER source
// files not present in this migration batch. As those controllers are migrated,
// replace the corresponding stubs with the real handler wiring (typically
// r.Mount or r.Method with a handler produced by a NewXxxHandler constructor).
func buildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Public routes owned by not-yet-migrated controllers.
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

	// Admin/user dashboard routes owned by not-yet-migrated controllers.
	r.Route("/user", func(r chi.Router) {
		r.Get("/index", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: UserDashboard")
		})
	})

	return r
}
