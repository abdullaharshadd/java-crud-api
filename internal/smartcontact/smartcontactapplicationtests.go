package smartcontact

// MIGRATION_NOTE: The Java source, SmartContactApplicationTests, was the default
// Spring Boot test class annotated with @SpringBootTest. It contained no test
// methods; its sole purpose was the convention-based "context loads" smoke
// test — Spring's test runner would boot the entire application context and
// fail if any bean could not be constructed or wired.
//
// Go has no runtime dependency-injection container or classpath scanning, so
// there is no application context to "load". The idiomatic equivalent of this
// smoke test is to exercise the explicit composition root (NewApp / Router,
// defined in smartcontactapplication.go) and assert that the wired-up HTTP
// router is successfully constructed and can serve requests. This verifies the
// same thing the Spring test did: that the repository -> service -> handler
// bean graph wires together without error.

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestContextLoads is the Go equivalent of the Spring Boot "context loads"
// smoke test. It builds the full application (composition root) and asserts
// that a usable HTTP router is produced. A nil router or a panic during
// construction is treated as a failure, mirroring Spring's behaviour of
// failing the test if the application context cannot be initialised.
func TestContextLoads(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil: application composition root failed to build")
	}

	router := app.Router()
	if router == nil {
		t.Fatal("App.Router returned nil: HTTP router failed to wire up")
	}
}

// TestRouterServesRequests is a lightweight smoke test that confirms the wired
// router can accept and dispatch an HTTP request without panicking. It does
// not assert on any specific business response — it only verifies that the
// transport layer is reachable, which is the Go analogue of Spring verifying
// that the servlet container and controllers were registered.
//
// MIGRATION_NOTE: Exact golden-file response assertions for each endpoint
// (JSON bodies, status codes, textContentType) should be added once Variant
// A/B behaviour is confirmed, per the migration debate notes. Those belong in
// the handler package's tests where the routes and their contracts live.
func TestRouterServesRequests(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil: application composition root failed to build")
	}

	router := app.Router()
	if router == nil {
		t.Fatal("App.Router returned nil: HTTP router failed to wire up")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	// The router must dispatch without panicking. We intentionally do not
	// assert on the status code here: the underlying repository may be
	// unconfigured in this smoke test, so any well-formed HTTP response
	// (including an error status) proves the transport wiring is intact.
	router.ServeHTTP(rec, req)

	if rec.Code == 0 {
		t.Fatal("router did not write any response: transport layer is not wired")
	}
}
