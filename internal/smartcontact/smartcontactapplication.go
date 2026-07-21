// Package smartcontact is the application bootstrap package for the
// SmartContact service. It is the Go equivalent of the Spring Boot
// entry-point class com.smartContact.SmartContactApplication.
//
// MIGRATION_NOTE: The original SmartContactApplication.java contained no
// business logic. It was pure framework glue: the @SpringBootApplication
// meta-annotation triggered component scanning of the com.smartContact
// package, dependency auto-wiring, configuration loading from
// application.properties/yml, and startup of an embedded Tomcat server on
// port 8080.
//
// In idiomatic Go there is no reflection-based component scanning or
// dependency-injection container. The equivalent responsibilities are
// split explicitly:
//
//   - Component scanning / bean wiring -> explicit constructor wiring in
//     buildRouter (and, for larger apps, a dedicated wire/DI setup).
//   - Embedded server startup          -> http.Server in cmd/server/main.go.
//   - Config loading (application.yml)  -> a config package (migrated
//     separately) read at startup; the default port 8080 is carried into
//     cmd/server/main.go.
//   - Route registration               -> buildRouter below, using chi.
//
// buildRouter constructs the fully-wired HTTP handler and is invoked from
// cmd/server/main.go. Concrete controller routes are wired here as their
// controllers are migrated.
package smartcontact

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs and returns the fully-wired HTTP handler for the
// SmartContact application. It is the composition root for HTTP routing:
// middleware is installed here and controller routes are registered as
// their controllers are migrated.
//
// MIGRATION_NOTE: This entry-point file (SmartContactApplication.java)
// itself owned NO routes — it was pure Spring bootstrap glue. Therefore
// every route below is a placeholder 200-OK stub for a controller that
// lives in a separate, not-yet-migrated file. As each controller is
// migrated, replace the corresponding stub with the real handler wiring
// (typically a NewXxxController(...).Register(r) call). The /healthz
// endpoint is the only route genuinely owned here and is fully functional.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	// Baseline middleware chain (replaces Spring's implicit request logging
	// and error handling).
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check — genuinely owned by this bootstrap file.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Admin-prefixed routes are grouped for parity with typical
	// SmartContact admin controllers. These are stubs until the owning
	// controllers are migrated.
	r.Route("/admin", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDashboard")
		})
	})

	return r
}
