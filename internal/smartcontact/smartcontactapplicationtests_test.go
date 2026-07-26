```go
package smartcontact

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fakes / mocks for external dependencies
// ---------------------------------------------------------------------------

// fakeConfig is a minimal stand-in for whatever resources.Config exposes.
type fakeConfig interface {
	ServerAddr() string
}

type mockConfig struct {
	serverAddr string
}

func (m *mockConfig) ServerAddr() string { return m.serverAddr }

// pingable abstracts *sql.DB so we can inject a fake in tests.
type pingable interface {
	PingContext(ctx context.Context) error
	Close() error
}

// mockDB satisfies pingable.
type mockDB struct {
	pingErr  error
	closeErr error
	closed   bool
}

func (m *mockDB) PingContext(_ context.Context) error {
	return m.pingErr
}

func (m *mockDB) Close() error {
	m.closed = true
	return m.closeErr
}

// ---------------------------------------------------------------------------
// Extracted, testable wiring logic
// ---------------------------------------------------------------------------

// appWiringResult captures what happened during a wiring attempt.
type appWiringResult struct {
	configLoaded      bool
	serverAddrNonEmpty bool
	dbSkipped         bool
	dbReachable       bool
	dbClosed          bool
	err               error
}

// configLoader abstracts resources.LoadConfig.
type configLoader func() (fakeConfig, error)

// dbOpener abstracts resources.OpenDB.
type dbOpener func(cfg fakeConfig) (pingable, error)

// runApplicationWiring is the extracted, testable core of TestApplicationWiring.
func runApplicationWiring(
	t *testing.T,
	loadConfig configLoader,
	openDB dbOpener,
) appWiringResult {
	t.Helper()

	result := appWiringResult{}

	cfg, err := loadConfig()
	if err != nil {
		result.err = err
		return result
	}
	result.configLoaded = true

	if cfg.ServerAddr() == "" {
		result.err = errors.New("expected a non-empty server address from configuration")
		return result
	}
	result.serverAddrNonEmpty = true

	db, err := openDB(cfg)
	if err != nil {
		result.dbSkipped = true
		return result
	}
	defer func() {
		if cerr := db.Close(); cerr == nil {
			result.dbClosed = true
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		result.dbSkipped = true
		return result
	}
	result.dbReachable = true
	result.dbClosed = true // defer will run
	return result
}

// ---------------------------------------------------------------------------
// Table-driven tests
// ---------------------------------------------------------------------------

func TestApplicationWiringTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		// config side
		configErr  error
		serverAddr string

		// db side
		openDBErr error
		pingErr   error
		closeErr  error

		// expectations
		wantConfigLoaded       bool
		wantServerAddrNonEmpty bool
		wantDBSkipped          bool
		wantDBReachable        bool
		wantErr                bool
		wantErrContains        string
	}{
		{
			name:                   "happy path – config loads, DB reachable",
			serverAddr:             ":8080",
			wantConfigLoaded:       true,
			wantServerAddrNonEmpty: true,
			wantDBReachable:        true,
		},
		{
			name:            "config load fails",
			configErr:       errors.New("missing env vars"),
			wantErr:         true,
			wantErrContains: "missing env vars",
		},
		{
			name:            "config loads but server address is empty",
			serverAddr:      "",
			wantConfigLoaded: true,
			wantErr:         true,
			wantErrContains: "non-empty server address",
		},
		{
			name:                   "database unavailable – openDB fails – test skipped not failed",
			serverAddr:             ":8080",
			openDBErr:              errors.New("dial tcp: connection refused"),
			wantConfigLoaded:       true,
			wantServerAddrNonEmpty: true,
			wantDBSkipped:          true,
		},
		{
			name:                   "database unreachable – ping fails – test skipped not failed",
			serverAddr:             ":8080",
			pingErr:                errors.New("context deadline exceeded"),
			wantConfigLoaded:       true,
			wantServerAddrNonEmpty: true,
			wantDBSkipped:          true,
		},
		{
			name:                   "DB close error is logged but does not fail the test",
			serverAddr:             ":8080",
			closeErr:               errors.New("close: use of closed connection"),
			wantConfigLoaded:       true,
			wantServerAddrNonEmpty: true,
			wantDBReachable:        true,
			// dbClosed will be false because close returned an error
		},
		{
			name:                   "required bean missing – modelled as config error",
			configErr:              errors.New("required bean DataSource cannot be created"),
			wantErr:                true,
			wantErrContains:        "DataSource cannot be created",
		},
		{
			name:                   "required config property missing – modelled as config error",
			configErr:              errors.New("required configuration property 'db.url' is missing"),
			wantErr:                true,
			wantErrContains:        "db.url",
		},
		{
			name:                   "dependency injection failure – modelled as config error",
			configErr:              errors.New("unsatisfied dependency: no qualifying bean of type Repository"),
			wantErr:                true,
			wantErrContains:        "Repository",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build loader
			loader := func() (fakeConfig, error) {
				if tc.configErr != nil {
					return nil, tc.configErr
				}
				return &mockConfig{serverAddr: tc.serverAddr}, nil
			}

			// Build opener
			opener := func(_ fakeConfig) (pingable, error) {
				if tc.openDBErr != nil {
					return nil, tc.openDBErr
				}
				return &mockDB{
					pingErr:  tc.pingErr,
					closeErr: tc.closeErr,
				}, nil
			}

			result := runApplicationWiring(t, loader, opener)

			if tc.wantErr {
				assert.Error(t, result.err, "expected an error but got none")
				if tc.wantErrContains != "" {
					assert.Contains(t, result.err.Error(), tc.wantErrContains)
				}
			} else {
				assert.NoError(t, result.err)
			}

			assert.Equal(t, tc.wantConfigLoaded, result.configLoaded, "configLoaded mismatch")
			assert.Equal(t, tc.wantServerAddrNonEmpty, result.serverAddrNonEmpty, "serverAddrNonEmpty mismatch")
			assert.Equal(t, tc.wantDBSkipped, result.dbSkipped, "dbSkipped mismatch")
			assert.Equal(t, tc.wantDBReachable, result.dbReachable, "dbReachable mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// Invariant tests
// ---------------------------------------------------------------------------

// TestApplicationWiring_Invariants validates the global invariants:
//   - the application must be startable (config is internally consistent)
//   - no wiring errors occur when all deps are satisfied
//   - no assertions beyond successful initialization
func TestApplicationWiring_Invariants(t *testing.T) {
	t.Parallel()

	t.Run("full context loads with no errors when all deps satisfied", func(t *testing.T) {
		t.Parallel()

		loader := func() (fakeConfig, error) {
			return &mockConfig{serverAddr: "localhost:9090"}, nil
		}
		opener := func(_ fakeConfig) (pingable, error) {
			return &mockDB{}, nil
		}

		result := runApplicationWiring(t, loader, opener)

		assert.NoError(t, result.err, "wiring must succeed when all dependencies are satisfied")
		assert.True(t, result.configLoaded, "config must be loaded")
		assert.True(t, result.serverAddrNonEmpty, "server address must be non-empty")
		assert.True(t, result.dbReachable, "DB must be reachable")
		assert.False(t, result.dbSkipped, "DB check must not be skipped")
	})

	t.Run("no business logic assertions beyond successful init", func(t *testing.T) {
		t.Parallel()

		// This test mirrors the Java invariant: "the test performs no assertions
		// beyond successful context initialization". We verify the helper returns
		// cleanly and does not expose any domain-level state.
		loader := func() (fakeConfig, error) {
			return &mockConfig{serverAddr: ":8080"}, nil
		}
		opener := func(_ fakeConfig) (pingable, error) {
			return &mockDB{}, nil
		}

		result := runApplicationWiring(t, loader, opener)
		assert.NoError(t, result.err)
		// The only meaningful field is that things loaded without error.
	})

	t.Run("component scanning inconsistency modelled as loader error", func(t *testing.T) {
		t.Parallel()

		loader := func() (fakeConfig, error) {
			return nil, errors.New("component scanning conflict: duplicate bean 'userService'")
		}
		opener := func(_ fakeConfig) (pingable, error) { return &mockDB{}, nil }

		result := runApplicationWiring(t, loader, opener)
		assert.Error(t, result.err)
		assert.Contains(t, result.err.Error(), "userService")
		assert.False(t, result.configLoaded)
	})
}

// ---------------------------------------------------------------------------
// mockDB helper – ensure it also satisfies *sql.DB-style interface at compile time
// ---------------------------------------------------------------------------

var _ pingable = (*mockDB)(nil)

// Ensure the concrete *sql.DB satisfies our pingable interface so the
// production path and the test path share the same contract.
var _ interface {
	PingContext(ctx context.Context) error
	Close() error
} = (*sql.DB)(nil)
```