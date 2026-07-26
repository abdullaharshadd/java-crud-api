```go
package resources_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/resources"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setEnv sets key=value for the duration of a test and restores the previous
// state via t.Cleanup.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// unsetEnv removes key for the duration of a test and restores it afterwards.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, prev)
		}
	})
}

// ---------------------------------------------------------------------------
// Constants / defaults
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	t.Run("DefaultServerPort is 8082", func(t *testing.T) {
		assert.Equal(t, 8082, resources.DefaultServerPort)
	})
	t.Run("DefaultDBHost is localhost", func(t *testing.T) {
		assert.Equal(t, "localhost", resources.DefaultDBHost)
	})
	t.Run("DefaultDBPort is 3306", func(t *testing.T) {
		assert.Equal(t, 3306, resources.DefaultDBPort)
	})
	t.Run("DefaultDBName is barcode", func(t *testing.T) {
		assert.Equal(t, "barcode", resources.DefaultDBName)
	})
	t.Run("DefaultDBUser is root", func(t *testing.T) {
		assert.Equal(t, "root", resources.DefaultDBUser)
	})
	t.Run("DefaultDBPassword is root", func(t *testing.T) {
		assert.Equal(t, "root", resources.DefaultDBPassword)
	})
}

// ---------------------------------------------------------------------------
// NewConfig – defaults (no env overrides)
// ---------------------------------------------------------------------------

func TestNewConfig_Defaults(t *testing.T) {
	// Ensure none of the recognised env vars are set.
	for _, key := range []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD"} {
		unsetEnv(t, key)
	}

	cfg, err := resources.NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, resources.DefaultServerPort, cfg.ServerPort,
		"server.port must default to 8082 (mirrors server.port=8082)")
	assert.Equal(t, resources.DefaultDBHost, cfg.DBHost,
		"DBHost must default to localhost")
	assert.Equal(t, resources.DefaultDBPort, cfg.DBPort,
		"DBPort must default to 3306")
	assert.Equal(t, resources.DefaultDBName, cfg.DBName,
		"DBName must default to barcode")
	assert.Equal(t, resources.DefaultDBUser, cfg.DBUser,
		"DBUser must default to root")
	assert.Equal(t, resources.DefaultDBPassword, cfg.DBPassword,
		"DBPassword must default to root")
}

// ---------------------------------------------------------------------------
// NewConfig – environment overrides (table-driven)
// ---------------------------------------------------------------------------

func TestNewConfig_EnvOverrides(t *testing.T) {
	// Ensure no leftover env from parent process.
	for _, key := range []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD"} {
		unsetEnv(t, key)
	}

	tests := []struct {
		name   string
		envs   map[string]string
		expect resources.Config
	}{
		{
			name: "override SERVER_PORT",
			envs: map[string]string{"SERVER_PORT": "9090"},
			expect: resources.Config{
				ServerPort: 9090,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
			},
		},
		{
			name: "override DB_HOST",
			envs: map[string]string{"DB_HOST": "db.example.com"},
			expect: resources.Config{
				ServerPort: resources.DefaultServerPort,
				DBHost:     "db.example.com",
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
			},
		},
		{
			name: "override DB_PORT",
			envs: map[string]string{"DB_PORT": "3307"},
			expect: resources.Config{
				ServerPort: resources.DefaultServerPort,
				DBHost:     resources.DefaultDBHost,
				DBPort:     3307,
				DBName:     resources.DefaultDBName,
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
			},
		},
		{
			name: "override DB_NAME",
			envs: map[string]string{"DB_NAME": "mydb"},
			expect: resources.Config{
				ServerPort: resources.DefaultServerPort,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     "mydb",
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
			},
		},
		{
			name: "override DB_USER",
			envs: map[string]string{"DB_USER": "admin"},
			expect: resources.Config{
				ServerPort: resources.DefaultServerPort,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
				DBUser:     "admin",
				DBPassword: resources.DefaultDBPassword,
			},
		},
		{
			name: "override DB_PASSWORD",
			envs: map[string]string{"DB_PASSWORD": "s3cr3t"},
			expect: resources.Config{
				ServerPort: resources.DefaultServerPort,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
				DBUser:     resources.DefaultDBUser,
				DBPassword: "s3cr3t",
			},
		},
		{
			name: "override all variables",
			envs: map[string]string{
				"SERVER_PORT": "8080",
				"DB_HOST":     "prod-db",
				"DB_PORT":     "3308",
				"DB_NAME":     "proddb",
				"DB_USER":     "svc",
				"DB_PASSWORD": "hunter2",
			},
			expect: resources.Config{
				ServerPort: 8080,
				DBHost:     "prod-db",
				DBPort:     3308,
				DBName:     "proddb",
				DBUser:     "svc",
				DBPassword: "hunter2",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				setEnv(t, k, v)
			}

			cfg, err := resources.NewConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.expect.ServerPort, cfg.ServerPort)
			assert.Equal(t, tc.expect.DBHost, cfg.DBHost)
			assert.Equal(t, tc.expect.DBPort, cfg.DBPort)
			assert.Equal(t, tc.expect.DBName, cfg.DBName)
			assert.Equal(t, tc.expect.DBUser, cfg.DBUser)
			assert.Equal(t, tc.expect.DBPassword, cfg.DBPassword)
		})
	}
}

// ---------------------------------------------------------------------------
// NewConfig – error cases for invalid numeric overrides
// ---------------------------------------------------------------------------

func TestNewConfig_InvalidNumericEnvVars(t *testing.T) {
	for _, key := range []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD"} {
		unsetEnv(t, key)
	}

	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr string
	}{
		{
			name:    "invalid SERVER_PORT returns error",
			envKey:  "SERVER_PORT",
			envVal:  "not-a-number",
			wantErr: "SERVER_PORT",
		},
		{
			name:    "SERVER_PORT with float returns error",
			envKey:  "SERVER_PORT",
			envVal:  "80.5",
			wantErr: "SERVER_PORT",
		},
		{
			name:    "invalid DB_PORT returns error",
			envKey:  "DB_PORT",
			envVal:  "abc",
			wantErr: "DB_PORT",
		},
		{
			name:    "DB_PORT with empty string returns error",
			envKey:  "DB_PORT",
			envVal:  "",
			wantErr: "DB_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.envKey, tc.envVal)

			cfg, err := resources.NewConfig()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tc.wantErr,
				"error message should reference the problematic env var")
		})
	}
}

// ---------------------------------------------------------------------------
// BuildDSN – table-driven
// ---------------------------------------------------------------------------

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name    string
		cfg     resources.Config
		wantDSN string
	}{
		{
			name: "default config produces correct DSN",
			cfg: resources.Config{
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
			},
			wantDSN: "root:root@tcp(localhost:3306)/barcode?parseTime=true",
		},
		{
			name: "custom credentials and host",
			cfg: resources.Config{
				DBUser:     "admin",
				DBPassword: "s3cr3t",
				DBHost:     "db.internal",
				DBPort:     3307,
				DBName:     "myapp",
			},
			wantDSN: "admin:s3cr3t@tcp(db.internal:3307)/myapp?parseTime=true",
		},
		{
			name: "empty password still forms valid DSN",
			cfg: resources.Config{
				DBUser:     "root",
				DBPassword: "",
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
			},
			wantDSN: "root:@tcp(localhost:3306)/barcode?parseTime=true",
		},
		{
			name: "non-standard port",
			cfg: resources.Config{
				DBUser:     "svc",
				DBPassword: "pass",
				DBHost:     "10.0.0.1",
				DBPort:     13306,
				DBName:     "analytics",
			},
			wantDSN: "svc:pass@tcp(10.0.0.1:13306)/analytics?parseTime=true",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.BuildDSN()
			assert.Equal(t, tc.wantDSN, got)
		})
	}
}

func TestBuildDSN_DoesNotContainJDBC(t *testing.T) {
	cfg := resources.Config{
		DBUser:     "root",
		DBPassword: "root",
		DBHost:     "localhost",
		DBPort:     3306,
		DBName:     "barcode",
	}
	dsn := cfg.BuildDSN()
	assert.NotContains(t, dsn, "jdbc:", "Go DSN must not contain JDBC URL scheme")
	assert.NotContains(t, dsn, "jdbc:mysql", "Go DSN must not contain JDBC URL scheme")
}

func TestBuildDSN_ContainsParseTime(t *testing.T) {
	cfg := resources.Config{
		DBUser:     "root",
		DBPassword: "root",
		DBHost:     "localhost",
		DBPort:     3306,
		DBName:     "barcode",
	}
	dsn := cfg.BuildDSN()
	assert.Contains(t, dsn, "parseTime=true",
		"DSN must include parseTime=true for time.Time scanning")
}

func TestBuildDSN_TCPDialFormat(t *testing.T) {
	cfg := resources.Config{
		DBUser:     "root",
		DBPassword: "root",
		DBHost:     "localhost",
		DBPort:     3306,
		DBName:     "barcode",
	}
	dsn := cfg.BuildDSN()
	assert.Contains(t, dsn, "@tcp(",
		"DSN must use the go-sql-driver/mysql tcp dial format")
}

// ---------------------------------------------------------------------------
// ListenAddr – table-driven
// ---------------------------------------------------------------------------

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name       string
		serverPort int
		wantAddr   string
	}{
		{
			name:       "default port 8082 produces :8082",
			serverPort: resources.DefaultServerPort,
			wantAddr:   ":8082",
		},
		{
			name:       "port 8080 produces :8080",
			serverPort: 8080,
			wantAddr:   ":8080",
		},
		{
			name:       "port 443 produces :443",
			serverPort: 443,
			wantAddr:   ":443",
		},
		{
			name:       "port 0 produces :0 (OS-assigned)",
			serverPort: 0,
			wantAddr:   ":0",
		},
		{
			name:       "high port 65535 produces correct addr",
			serverPort: 65535,
			wantAddr:   ":65535",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &resources.Config{ServerPort: tc.serverPort}
			assert.Equal(t, tc.wantAddr, cfg.ListenAddr())
		})
	}
}

// ---------------------------------------------------------------------------
// server.port behavioural spec: HTTP server binds on port 8082
// ---------------------------------------------------------------------------

func TestServerPort_BindsOn8082_WithHttptest(t *testing.T) {
	// Use httptest.NewServer to verify the handler is reachable. Because
	// httptest picks its own port, we separately verify that ListenAddr()
	// produces 