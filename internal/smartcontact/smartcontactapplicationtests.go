package smartcontact

// MIGRATION_NOTE: The Java source (SmartContactApplicationTests) was the
// default Spring Boot generated smoke test annotated with @SpringBootTest.
// Its sole purpose was to verify that the Spring ApplicationContext could
// bootstrap successfully (component scanning, auto-configuration, embedded
// server wiring). It contained no assertions of its own — a clean context
// load counted as a pass.
//
// Go has no ApplicationContext to boot, so the idiomatic equivalent is a
// smoke test that exercises the same explicit dependency-injection wiring
// the real application performs: build the HTTP router from a Config +
// UserService and confirm it is non-nil. This mirrors
// smartcontactapplication.go's BuildRouterWith without touching a live
// database, which keeps the check hermetic and fast.
//
// Because this is a test, the real logic lives in the _test.go file below.
// This non-test file exists only to satisfy the requested target path with
// compilable, meaningful documentation of the migration decision; the
// actual smoke test is emitted alongside it.

// smokeTestName documents the origin of the migrated smoke test. It is a
// trivial exported-free constant kept so this file carries real code rather
// than only comments.
const smokeTestName = "SmartContactApplicationContextLoads"
