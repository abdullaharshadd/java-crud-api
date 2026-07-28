// Package smartcontact is the composition root for the SmartContact
// application. It replaces the Spring Boot entry point
// (com.smartContact.SmartContactApplication) which used @SpringBootApplication
// to trigger component scanning and auto-configuration.
//
// MIGRATION_NOTE: Go has no equivalent of Spring's @ComponentScan /
// @EnableAutoConfiguration. Instead of magic wiring, dependencies are wired
// explicitly. The HTTP router assembly lives here in buildRouter(); the actual
// process bootstrap (config loading, DB connection, graceful shutdown) lives in
// cmd/server/main.go, which calls buildRouter() to obtain the http.Handler.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the
// SmartContact application. It is invoked by cmd/server/main.go.
//
// MIGRATION_NOTE: The original Spring class contained no route definitions or
// business logic of its own — routing was discovered via component scanning of
// @Controller/@RestController beans in other packages. Those controllers are
// migrated in their own files. Because none of them are present in this
// migration batch, their routes are registered here as explicit 200-OK stubs
// so the server remains runnable; each stub must be replaced by wiring the real
// migrated handler when its controller file is migrated.
func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// MIGRATION_NOTE: No routes were owned by SmartContactApplication.java
	// itself (it was purely the bootstrap class). As controller files
	// (e.g. HomeController, UserController) are migrated, register their real
	// handlers below and under the /admin group as appropriate.
	r.Route("/admin", func(r chi.Router) {
		// Placeholder for admin-prefixed routes owned by not-yet-migrated
		// controller files. Replace with real handlers as they are migrated.
	})

	return r
}
