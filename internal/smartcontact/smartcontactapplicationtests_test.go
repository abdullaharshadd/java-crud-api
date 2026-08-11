```go
package smartcontact

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/smartContact/internal/smartcontact/service"
)

// ----------------------------------------------------------------------------
// Minimal mock driver so we can exercise the wiring without a real Postgres.
// ----------------------------------------------------------------------------

type mockDriver struct{ failPing bool }

func (d *mockDriver) Open(_ string) (driver.Conn, error) {
	if d.failPing {
		return nil, fmt.Errorf("mock: connection refused")
	}
	return &mockConn{}, nil
}

type mockConn struct{}

func (*mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (*mockConn) Close() error                              { return nil }
func (*mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }

type mockStmt struct{}

func (*mockStmt) Close() error                                    { return nil }
func (*mockStmt) NumInput() int                                   { return 0 }
func (*mockStmt) Exec(_ []driver.Value) (driver.Result, error)   { return driver.ResultNoRows, nil }
func (*mockStmt) Query(_ []driver.Value) (driver.Rows, error)    { return &mockRows{}, nil }

type mockRows struct{ done bool }

func (r *mockRows) Columns() []string          { return nil }
func (r *mockRows) Close() error               { return nil }
func (r *mockRows) Next(_ []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	return io.EOF
}

type mockTx struct{}

func (*mockTx) Commit() error   { return nil }
func (*mockTx) Rollback() error { return nil }

// ----------------------------------------------------------------------------
// mockPinger lets us test the ping-and-skip path without a real DB.
// ----------------------------------------------------------------------------

type mockPinger struct {
	db      *sql.DB
	pingErr error
}

// ----------------------------------------------------------------------------
// Interfaces that capture the wiring contracts (used only inside tests).
// ----------------------------------------------------------------------------

type userDao interface {
	// At minimum the repository type must be non-nil; full method coverage
	// lives in the repository package's own tests.
}

type userSvc interface{}
type userHndlr interface {
	RegisterRoutes(r chi.Router)
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func init() {
	// Register a deterministic mock driver name so tests that need a *sql.DB
	// can obtain one without a real Postgres.
	sql.Register("mockpg", &mockDriver{})
}

// newMockDB returns a *sql.DB backed by the mock driver.
func newMockDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("mockpg", "mock://localhost/test")
	require.NoError(t, err, "sql.Open with mock driver must not fail")
	return db
}

// ----------------------------------------------------------------------------
// Table-driven tests
// ----------------------------------------------------------------------------

// TestContextLoads_Wiring validates the full dependency-graph wiring with a
// mock DB, so it runs without Postgres. This is the primary "context loads"
// smoke test.
func TestContextLoads_Wiring(t *testing.T) {
	tests := []struct {
		name        string
		buildDB     func(t *testing.T) *sql.DB
		wantRepo    bool
		wantSvc     bool
		wantHandler bool
		wantRoute   bool
	}{
		{
			name:        "all layers wire successfully with valid DB",
			buildDB:     newMockDB,
			wantRepo:    true,
			wantSvc:     true,
			wantHandler: true,
			wantRoute:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := tc.buildDB(t)
			defer db.Close()

			// Repository layer
			repo := repository.NewUserDao(db)
			if tc.wantRepo {
				assert.NotNil(t, repo, "NewUserDao must return a non-nil repository")
			} else {
				assert.Nil(t, repo)
			}

			// Service layer
			svc := service.NewUserService(repo)
			if tc.wantSvc {
				assert.NotNil(t, svc, "NewUserService must return a non-nil service")
			} else {
				assert.Nil(t, svc)
			}

			// Handler layer
			h := handler.NewUserHandler(svc)
			if tc.wantHandler {
				assert.NotNil(t, h, "NewUserHandler must return a non-nil handler")
			} else {
				assert.Nil(t, h)
			}

			// Route registration
			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tc.wantRoute {
				assert.NotEqual(t, http.StatusNotFound, rec.Code,
					"GET /api/users must be registered; got 404 which means the route is absent")
			}
		})
	}
}

// TestContextLoads_RepositoryNotNil specifically guards the invariant that
// the repository is never nil when a valid DB is supplied.
func TestContextLoads_RepositoryNotNil(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			name:    "NewUserDao with valid db returns non-nil",
			wantNil: false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB(t)
			defer db.Close()
			repo := repository.NewUserDao(db)
			if tc.wantNil {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
			}
		})
	}
}

// TestContextLoads_ServiceNotNil guards the invariant that the service is
// never nil when a valid repository is supplied.
func TestContextLoads_ServiceNotNil(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			name:    "NewUserService with valid repo returns non-nil",
			wantNil: false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB(t)
			defer db.Close()
			repo := repository.NewUserDao(db)
			require.NotNil(t, repo)

			svc := service.NewUserService(repo)
			if tc.wantNil {
				assert.Nil(t, svc)
			} else {
				assert.NotNil(t, svc)
			}
		})
	}
}

// TestContextLoads_HandlerNotNil guards the invariant that the handler is
// never nil when a valid service is supplied.
func TestContextLoads_HandlerNotNil(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			name:    "NewUserHandler with valid service returns non-nil",
			wantNil: false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB(t)
			defer db.Close()
			repo := repository.NewUserDao(db)
			require.NotNil(t, repo)
			svc := service.NewUserService(repo)
			require.NotNil(t, svc)

			h := handler.NewUserHandler(svc)
			if tc.wantNil {
				assert.Nil(t, h)
			} else {
				assert.NotNil(t, h)
			}
		})
	}
}

// TestContextLoads_RouteRegistration checks that RegisterRoutes actually
// registers HTTP routes on the chi router (i.e. the router is not empty).
func TestContextLoads_RouteRegistration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantRegistered bool // true => must NOT be 404
	}{
		{
			name:           "GET /api/users route is registered",
			method:         http.MethodGet,
			path:           "/api/users",
			wantRegistered: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB(t)
			defer db.Close()
			repo := repository.NewUserDao(db)
			require.NotNil(t, repo)
			svc := service.NewUserService(repo)
			require.NotNil(t, svc)
			h := handler.NewUserHandler(svc)
			require.NotNil(t, h)

			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tc.wantRegistered {
				assert.NotEqual(t, http.StatusNotFound, rec.Code,
					"route %s %s must be registered on the router", tc.method, tc.path)
			} else {
				assert.Equal(t, http.StatusNotFound, rec.Code,
					"route %s %s must NOT be registered", tc.method, tc.path)
			}
		})
	}
}

// TestContextLoads_NoExplicitAssertionBeyondWiring validates the invariant
// that the smoke test succeeds purely through successful construction. No
// business logic is asserted -- mirroring the Spring @SpringBootTest contract.
func TestContextLoads_NoExplicitAssertionBeyondWiring(t *testing.T) {
	// If we reach this line, every constructor completed without panic or
	// error. The table below documents the invariant in structured form.
	tests := []struct {
		name            string
		constructionErr error // nil => wiring succeeded
	}{
		{
			name:            "all constructors complete without error",
			constructionErr: nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB(t)
			defer db.Close()

			var constructionErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						constructionErr = fmt.Errorf("panic during wiring: %v", r)
					}
				}()
				repo := repository.NewUserDao(db)
				svc := service.NewUserService(repo)
				h := handler.NewUserHandler(svc)
				r := chi.NewRouter()
				h.RegisterRoutes(r)
			}()

			assert.Equal(t, tc.constructionErr, constructionErr,
				"wiring must complete without error or panic")
		})
	}
}

// TestOpenTestDB_SkipsWhenDBUnreachable verifies that openTestDB calls
// t.Skip (and does not t.Fatal) when Postgres is not available. Because we
// cannot intercept t.Skip from outside, we use a fakeT that records the call.
func TestOpenTestDB_SkipsWhenDBUnreachable(t *testing.T) {
	tests := []struct {
		name       string
		dsn        string
		wantSkip   bool
		wantFatal  bool
	}{
		{
			name:      "unreachable host causes skip",
			dsn:       "postgres://postgres:postgres@127.0.0.1:19999/nonexistent?sslmode=disable&connect_timeout=1",
			wantSkip:  true,
			wantFatal: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeT{}

			// Call openTestDB via a shim that accepts an alternate DSN so we
			// don't hard-code localhost:5432 in the assertion.
			openTestDBWithDSN(ft, tc.dsn)

			assert.Equal(t, tc.wantSkip, ft.skipped,
				"openTestDB must skip when DB is unreachable")
			assert.Equal(t, tc.wantFatal, ft.fataled,
				"openTestDB must not fatal when DB is merely unreachable")
		})
	}
}

// TestContextLoads_WithContext verifies that the smoke test respects context
// deadlines (i.e. the wiring itself does not block indefinitely).
func TestContextLoads_WithContext(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		expectTimeout bool
	}{
		{
			name:          "wiring completes well within 10-second deadline",
			timeout:       10 * time.Second,
			expectTimeout: false,
		},
		{
			name:          "wiring completes within a tight 5-second deadline",
			timeout:       5 * time.Second,
			expectTimeout: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer close(done)
				db := newMockDB(t)
				defer db.Close()
				repo := repository.NewUserDao(db)
				svc := service.NewUserService(repo)
				h := handler.NewUserHandler(svc)
				r := chi.NewRouter()
				h.RegisterRoutes(r)

				req := httptest.NewRequest(http.MethodGet, "/api/users", nil).WithContext(ctx)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
			}()

			select {
			case <-ctx.Done():
				if !tc.expectTimeout {
					t.Errorf("context timed out before wiring completed: %v", ctx.Err())
				}
			case <-done:
				if tc.expectTimeout {
					t.Error("expected timeout but wiring completed")
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// fakeT + helpers
// ----------------------------------------------------------------------------

// fakeT is a minimal stand-in for *testing.T that captures Skip/Fatal calls
// so we can assert on them without actually aborting the parent test.
type fakeT struct {
	skipped bool
	fataled bool
	logs    []string
}

func (f *fakeT) Helper() {}
func (f *fakeT) Skipf(format string, args ...interface{}) {
	f.skipped = true
}
func (f *fakeT) Fatalf(format string, args ...interface{}) {
	f.fataled = true
}
func (f *fakeT) Logf(format string, args ...interface{}) {
	f.logs = append(f.logs, fmt