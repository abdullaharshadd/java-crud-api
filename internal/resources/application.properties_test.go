```go
package resources_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/resources"
)

// ---------------------------------------------------------------------------
// Minimal in-process fake SQL driver used to test OpenDB without a real DB.
// ---------------------------------------------------------------------------

type fakeDriver struct {
	shouldFailOpen bool
	shouldFailPing bool
}

type fakeConn struct{ failPing bool }
type fakeStmt struct{}
type fakeTx struct{}
type fakeResult struct{}
type fakeRows struct{ closed bool }

func (f *fakeDriver) Open(name string) (driver.Conn, error) {
	if f.shouldFailOpen {
		return nil, errors.New("fake: open failed")
	}
	return &fakeConn{failPing: f.shouldFailPing}, nil
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) { return &fakeStmt{}, nil }
func (c *fakeConn) Close() error                              { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)                 { return &fakeTx{}, nil }

// Implement driver.Pinger so the fake conn can control Ping behaviour.
func (c *fakeConn) Ping(ctx context.Context) error {
	if c.failPing {
		return errors.New("fake: ping failed")
	}
	return nil
}

func (s *fakeStmt) Close() error                                    { return nil }
func (s *fakeStmt) NumInput() int                                   { return 0 }
func (s *fakeStmt) Exec(_ []driver.Value) (driver.Result, error)    { return &fakeResult{}, nil }
func (s *fakeStmt) Query(_ []driver.Value) (driver.Rows, error)     { return &fakeRows{}, nil }
func (t *fakeTx) Commit() error                                     { return nil }
func (t *fakeTx) Rollback() error                                   { return nil }
func (r *fakeResult) LastInsertId() (int64, error)                  { return 0, nil }
func (r *fakeResult) RowsAffected() (int64, error)                  { return 0, nil }
func (r *fakeRows) Columns() []string                               { return nil }
func (r *fakeRows) Close() error                                    { return nil }
func (r *fakeRows) Next(_ []driver.Value) error                     { return sql.ErrNoRows }

// registerDriver registers a named fake driver only once.
var registeredDrivers = map[string]bool{}

func registerFakeDriver(name string, d driver.Driver) {
	if !registeredDrivers[name] {
		sql.Register(name, d)
		registeredDrivers[name] = true
	}
}

func init() {
	registerFakeDriver("fake_ok", &fakeDriver{})
	registerFakeDriver("fake_ping_fail", &fakeDriver{shouldFailPing: true})
}

// ---------------------------------------------------------------------------
// Helper: set env vars and restore on cleanup
// ---------------------------------------------------------------------------

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	orig, existed := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, orig)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	orig, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, orig)
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for LoadConfig – defaults
// ---------------------------------------------------------------------------

func TestLoadConfig_Defaults(t *testing.T) {
	// Unset every env var that LoadConfig reads so we get pure defaults.
	keys := []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE"}
	for _, k := range keys {
		unsetEnv(t, k)
	}

	cfg, err := resources.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 8082, cfg.ServerPort, "default server port must be 8082")
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "barcode", cfg.DBName)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "postgres", cfg.DBPassword)
	assert.Equal(t, "disable", cfg.DBSSLMode)
}

// ---------------------------------------------------------------------------
// Tests for LoadConfig – env var overrides
// ---------------------------------------------------------------------------

func TestLoadConfig_EnvOverrides(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		check   func(*testing.T, *resources.Config)
	}{
		{
			name:   "SERVER_PORT override",
			envKey: "SERVER_PORT",
			envVal: "9090",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, 9090, c.ServerPort)
			},
		},
		{
			name:   "DB_HOST override",
			envKey: "DB_HOST",
			envVal: "db.example.com",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, "db.example.com", c.DBHost)
			},
		},
		{
			name:   "DB_PORT override",
			envKey: "DB_PORT",
			envVal: "5433",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, 5433, c.DBPort)
			},
		},
		{
			name:   "DB_NAME override",
			envKey: "DB_NAME",
			envVal: "mydb",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, "mydb", c.DBName)
			},
		},
		{
			name:   "DB_USER override",
			envKey: "DB_USER",
			envVal: "admin",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, "admin", c.DBUser)
			},
		},
		{
			name:   "DB_PASSWORD override",
			envKey: "DB_PASSWORD",
			envVal: "secret",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, "secret", c.DBPassword)
			},
		},
		{
			name:   "DB_SSLMODE override",
			envKey: "DB_SSLMODE",
			envVal: "require",
			check: func(t *testing.T, c *resources.Config) {
				assert.Equal(t, "require", c.DBSSLMode)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Unset all env vars, then set the target one.
			keys := []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE"}
			for _, k := range keys {
				unsetEnv(t, k)
			}
			setEnv(t, tc.envKey, tc.envVal)

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			tc.check(t, cfg)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for LoadConfig – invalid numeric env vars
// ---------------------------------------------------------------------------

func TestLoadConfig_InvalidNumericEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envVal      string
		errContains string
	}{
		{
			name:        "invalid SERVER_PORT",
			envKey:      "SERVER_PORT",
			envVal:      "not-a-number",
			errContains: "SERVER_PORT",
		},
		{
			name:        "invalid DB_PORT",
			envKey:      "DB_PORT",
			envVal:      "abc",
			errContains: "DB_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			keys := []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE"}
			for _, k := range keys {
				unsetEnv(t, k)
			}
			setEnv(t, tc.envKey, tc.envVal)

			cfg, err := resources.LoadConfig()
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for Config.ServerAddr
// ---------------------------------------------------------------------------

func TestConfig_ServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected string
	}{
		{
			name:     "default port 8082",
			port:     8082,
			expected: ":8082",
		},
		{
			name:     "port 9090",
			port:     9090,
			expected: ":9090",
		},
		{
			name:     "port 80",
			port:     80,
			expected: ":80",
		},
		{
			name:     "port 443",
			port:     443,
			expected: ":443",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &resources.Config{ServerPort: tc.port}
			assert.Equal(t, tc.expected, cfg.ServerAddr())
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for Config.DSN
// ---------------------------------------------------------------------------

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      resources.Config
		contains []string
	}{
		{
			name: "default values produce valid DSN",
			cfg: resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			contains: []string{
				"host=localhost",
				"port=5432",
				"user=postgres",
				"password=postgres",
				"dbname=barcode",
				"sslmode=disable",
			},
		},
		{
			name: "custom values reflected in DSN",
			cfg: resources.Config{
				DBHost:     "db.prod.example.com",
				DBPort:     5433,
				DBUser:     "admin",
				DBPassword: "hunter2",
				DBName:     "production",
				DBSSLMode:  "require",
			},
			contains: []string{
				"host=db.prod.example.com",
				"port=5433",
				"user=admin",
				"password=hunter2",
				"dbname=production",
				"sslmode=require",
			},
		},
		{
			name: "barcode database name is preserved (migration from MySQL)",
			cfg: resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			contains: []string{"dbname=barcode"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dsn := tc.cfg.DSN()
			for _, substr := range tc.contains {
				assert.Contains(t, dsn, substr, "DSN should contain %q", substr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for Config.OpenDB
// ---------------------------------------------------------------------------

func TestConfig_OpenDB(t *testing.T) {
	tests := []struct {
		name        string
		driverName  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful open and ping",
			driverName: "fake_ok",
			wantErr:    false,
		},
		{
			name:        "ping failure surfaces error",
			driverName:  "fake_ping_fail",
			wantErr:     true,
			errContains: "pinging database",
		},
		{
			name:        "unregistered driver returns error",
			driverName:  "nonexistent_driver_xyz",
			wantErr:     true,
			errContains: "opening database",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db, err := cfg.OpenDB(ctx, tc.driverName)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
				// Clean up.
				_ = db.Close()
			}
		})
	}
}

func TestConfig_OpenDB_ContextCancelled(t *testing.T) {
	cfg := &resources.Config{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "barcode",
		DBSSLMode:  "disable",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// With a cancelled context and a real driver that requires network, we
	// expect an error or a successful local fake. Since fake_ok does not
	// actually block on context, we use the ping-fail driver to verify
	// error propagation.
	db, err := cfg.OpenDB(ctx, "fake_ping_fail")
	assert.Error(t, err)
	assert.Nil(t, db)
}

// ---------------------------------------------------------------------------
// Tests for RunMigrations – deliberate no-op placeholder
// ---------------------------------------------------------------------------

func TestRunMigrations_IsNoOp(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		db   *sql.DB
	}{
		{
			name: "nil db is accepted (placeholder)",
			ctx:  context.Background(),
			db:   nil,
		},
		{
			name: "background context with nil db