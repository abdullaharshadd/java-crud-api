package smartcontact

// MIGRATION_NOTE: The Java source (SmartContactApplicationTests) was the default
// Spring Boot integration test class. It carried only the @SpringBootTest
// annotation and contained no test methods. Its sole purpose was Spring's
// implicit "contextLoads" smoke check: Spring Boot bootstraps the entire
// application context (component scanning, auto-configuration, bean wiring,
// datasource initialization) and the test passes if that startup succeeds.
//
// MIGRATION_NOTE: Go has no equivalent runtime context to "load". The Java
// composition-root magic (SpringApplication.run) is replaced by explicit,
// compile-time dependency injection in smartcontactapplication.go's
// buildRouter(). The Go analogue of "the context loads" is therefore "the
// wiring code constructs a non-nil router without panicking", which is exactly
// what this smoke test asserts.
//
// MIGRATION_NOTE: buildRouter is an unexported function in this same package,
// so this test lives in-package (package smartcontact, not smartcontact_test)
// to exercise it. If buildRouter's signature differs, adjust the call below;
// this is the single point a human should review.

import "testing"

// TestContextLoads is the Go equivalent of Spring Boot's implicit contextLoads
// smoke test. It verifies that the application's composition root wires all
// layers together and produces a usable HTTP router without failing.
func TestContextLoads(t *testing.T) {
	router := buildRouter()
	if router == nil {
		t.Fatal("buildRouter() returned nil; application router failed to initialize")
	}
}
