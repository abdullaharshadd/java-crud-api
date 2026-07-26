package smartcontact

// This file corresponds to the original Java test
// com.smartContact.SmartContactApplicationTests, the default Spring Boot
// auto-generated smoke test whose single purpose was to verify that the
// application's Spring context loads successfully (the conventional
// contextLoads() test).
//
// MIGRATION_NOTE: The Java source used @SpringBootTest to bootstrap the entire
// Spring application context (component scanning, dependency injection,
// DataSource/JPA auto-configuration, etc.) and confirm it initializes without
// error. Go has no equivalent framework-managed application context — wiring is
// explicit via constructor functions (NewXxx) and configuration loaded from the
// environment. There is therefore no "context" to load and no direct behavioral
// equivalent of contextLoads().
//
// The idiomatic Go analogue of this smoke test is to verify that the
// application's dependencies can be constructed and its configuration loaded
// without error. That is expressed below as a Go test (TestApplicationWiring)
// rather than a production source file, and it lives in this package so it can
// exercise the real wiring path.
//
// MIGRATION_NOTE: This test is intentionally minimal and integration-oriented.
// Per the migration debate it should be finalized last, once all components are
// stable. It reads configuration via resources.LoadConfig and, when a database
// is reachable, verifies a connection can be opened. When no database is
// configured/reachable the DB portion is skipped rather than failing, so the
// test can run in environments without external dependencies.
//
// This file is a Go test file (build-time _test.go semantics): rename to
// smartcontactapplicationtests_test.go so `go test` picks it up. The target
// path was fixed by the migration harness, so the doc above documents that
// requirement for manual review.

import (
	"context"
	"testing"
	"time"

	"github.com/smartContact/internal/resources"
)

// TestApplicationWiring is the Go equivalent of the Spring Boot contextLoads
// smoke test. It verifies that application configuration can be loaded and,
// when a database is available, that a connection can be established. It does
// not assert on any business logic; its sole purpose is to confirm the
// application's core wiring initializes without error.
func TestApplicationWiring(t *testing.T) {
	cfg, err := resources.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	if cfg.ServerAddr() == "" {
		t.Fatal("expected a non-empty server address from configuration")
	}

	// Attempt to open and ping the database. In environments without a
	// reachable database this is skipped rather than failed, mirroring the
	// intent of a lightweight smoke test that should not depend on external
	// infrastructure being present.
	db, err := resources.OpenDB(cfg)
	if err != nil {
		t.Skipf("database not available, skipping connectivity check: %v", err)
		return
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			t.Logf("error closing database: %v", cerr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("database not reachable, skipping connectivity check: %v", err)
	}
}
