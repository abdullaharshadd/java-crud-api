package smartcontact

// MIGRATION_NOTE: The Java source SmartContactApplicationTests.java was the
// default Spring Boot generated test:
//
//	@SpringBootTest
//	class SmartContactApplicationTests {
//	}
//
// It contained no real assertions — the sole implicit test was that Spring's
// application context could be built (all @Component/@Service/@Repository beans
// wired, DataSource connected). Go has no equivalent "context load" concept, so
// a direct translation would be an empty, valueless test.
//
// Per the migration debate (Changes 9, 13, 15) this file is replaced by a real
// MySQL-→-PostgreSQL integration test that exercises the repository layer against
// a live database, covering:
//   - FindByName ordering,
//   - affected-rows-as-not-found semantics for Update and DeleteByID,
//   - DB-assigned identity on create (RETURNING id, Postgres dialect),
//   - full response-field-coverage on ToResponse, excluding the password field.
//
// The test is guarded by an integration build tag AND an env var so `go test ./...`
// stays fast and hermetic; run it with:
//
//	go test -tags=integration -run TestUserDao ./internal/smartcontact/...
//
// with SMARTCONTACT_TEST_DSN pointing at a disposable PostgreSQL database.

// This file intentionally contains only the MIGRATION_NOTE above. The runnable
// integration test lives in the tagged test file below
// (smartcontactapplicationtests_integration_test.go) so that it participates in
// `go test` rather than compiling into the package binary. See that file for the
// actual assertions.
