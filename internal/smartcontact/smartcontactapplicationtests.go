package smartcontact

// MIGRATION_NOTE: The Java source (SmartContactApplicationTests) was the default
// Spring Boot generated test class annotated with @SpringBootTest. Its only
// purpose was the implicit "contextLoads" smoke test: Spring Boot bootstrapped
// the entire application context (component scanning, auto-configuration,
// DataSource wiring, embedded server) and the test passed if that succeeded
// without throwing. There was no explicit assertion or business logic.
//
// Go has no application context to bootstrap and no auto-configuration magic,
// so the faithful idiomatic equivalent is a smoke test that verifies the
// composition root can wire together the persistence, service, and HTTP handler
// layers into a usable http.Handler. buildRouter (defined in
// smartcontactapplication.go) is the Go analogue of the Spring context: if it
// returns a non-nil handler without error, the "context" loads successfully.
//
// MANUAL REVIEW: This mirrors the original test's intent (a wiring smoke test)
// but deliberately avoids a real database. Per the migration debate, the true
// end-to-end integration test verifying NULL handling across create and read
// paths should be added last, against a real or containerized PostgreSQL
// instance (e.g. via testcontainers-go). That is out of scope for this
// context-loads smoke test.

import (
	"testing"
)

// TestContextLoads is the Go equivalent of Spring Boot's implicit
// "contextLoads" smoke test. It verifies that the application's composition
// root (buildRouter) produces a non-nil http.Handler, which is the closest
// analogue to "the Spring application context started successfully".
//
// MIGRATION_NOTE: buildRouter is expected to be an unexported helper in
// smartcontactapplication.go that performs the manual constructor injection
// described in that file. If its signature differs (e.g. it requires a
// *sql.DB or *resources.Config, or returns an error), adjust the call below
// accordingly. We keep the dependency surface minimal here because a smoke
// test should not require external resources such as a live database.
func TestContextLoads(t *testing.T) {
	handler := buildRouter()
	if handler == nil {
		t.Fatal("buildRouter returned a nil handler; application wiring failed to load")
	}
}
