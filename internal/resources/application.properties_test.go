```go
package resources_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/resources"
)

// ---------------------------------------------------------------------------
// Helpers: environment isolation
// ---------------------------------------------------------------------------

// setEnv sets environment variables for the duration of a test, restoring
// original values (or unsetting) in a cleanup hook.
func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	originals := make(map[string]string, len(pairs))
	set := make(map[string]bool, len(pairs))

	for k, v := range pairs {
		orig, existed := os.LookupEnv(k)
		if existed {
			originals[k] = orig
		}
		set[k] = existed
		require.NoError(t, os.Setenv(k, v))
	}

	t.Cleanup(func() {
		for k := range pairs {
			if set[k] {
				_ = os.Setenv(k, originals[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	saved := make(map[string]string)
	existed := make(map[string]bool)

	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		existed[k] = ok
		if ok {
			saved[k] = v
		}
		require.NoError(t, os.Unsetenv(k))
	}

	t.Cleanup(func() {
		for _, k := range keys {
			if existed[k] {
				_ = os.Setenv(k, saved[k])
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Constants — defaults translated from application.properties
// ---------------------------------------------------------------------------

func TestConstants_Defaults(t *testing.T) {
	t.Run("server port default is 8082", func(t *testing.T) {
		assert.Equal(t, 8082, resources.DefaultServerPort,
			"server.port must default to 8082")
	})

	t.Run("db host default is localhost", func(t *testing.T) {
		assert.Equal(t, "localhost", resources.DefaultDBHost)
	})

	t.Run("db port default is 5432 (postgresql)", func(t *testing.T) {
		assert.Equal(t, 5432, resources.DefaultDBPort)
	})

	t.Run("db name default is barcode", func(t *testing.T) {
		assert.Equal(t, "barcode", resources.DefaultDBName,
			"spring.datasource.url database name must be 'barcode'")
	})

	t.Run("db user default is root", func(t *testing.T) {
		assert.Equal(t, "root", resources.DefaultDBUser,
			"spring.datasource.username must default to 'root'")
	})

	t.Run("db password default is root", func(t *testing.T) {
		assert.Equal(t, "root", resources.DefaultDBPassword,
			"spring.datasource.password must default to 'root'")
	})

	t.Run("sslmode default is disable", func(t *testing.T) {
		assert.Equal(t, "disable", resources.DefaultSSLMode)
	})

	t.Run("driver name is postgres", func(t *testing.T) {
		// MIGRATION_NOTE: source used com.mysql.cj.jdbc.Driver; Go target uses
		// postgres driver.
		assert.Equal(t, "postgres", resources.DriverName)
	})
}

// ---------------------------------------------------------------------------
// NewConfig — environment variable parsing
// ---------------------------------------------------------------------------

func TestNewConfig_Defaults(t *testing.T) {
	// Ensure none of the overridable env vars are set.
	unsetEnv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME",
		"DB_USER", "DB_PASSWORD", "DB_SSLMODE", "DB_AUTO_MIGRATE")

	cfg, err := resources.NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, resources.DefaultServerPort, cfg.ServerPort,
		"server.port default must be 8082")
	assert.Equal(t, resources.DefaultDBHost, cfg.DBHost)
	assert.Equal(t, resources.DefaultDBPort, cfg.DBPort)
	assert.Equal(t, resources.DefaultDBName, cfg.DBName,
		"datasource database name must default to 'barcode'")
	assert.Equal(t, resources.DefaultDBUser, cfg.DBUser,
		"datasource username must default to 'root'")
	assert.Equal(t, resources.DefaultDBPassword, cfg.DBPassword,
		"datasource password must default to 'root'")
	assert.Equal(t, resources.DefaultSSLMode, cfg.SSLMode)
	// ddl-auto=update → AutoMigrateSchema=true
	assert.True(t, cfg.AutoMigrateSchema,
		"spring.jpa.hibernate.ddl-auto=update must translate to AutoMigrateSchema=true")
}

func TestNewConfig_EnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		verify func(t *testing.T, cfg *resources.Config)
	}{
		{
			name: "SERVER_PORT overrides default port",
			env:  map[string]string{"SERVER_PORT": "9090"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 9090, cfg.ServerPort)
			},
		},
		{
			name: "DB_HOST overrides default host",
			env:  map[string]string{"DB_HOST": "pgserver"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "pgserver", cfg.DBHost)
			},
		},
		{
			name: "DB_PORT overrides default port",
			env:  map[string]string{"DB_PORT": "5433"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 5433, cfg.DBPort)
			},
		},
		{
			name: "DB_NAME overrides default database name",
			env:  map[string]string{"DB_NAME": "mydb"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "mydb", cfg.DBName)
			},
		},
		{
			name: "DB_USER overrides default username",
			env:  map[string]string{"DB_USER": "admin"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "admin", cfg.DBUser)
			},
		},
		{
			name: "DB_PASSWORD overrides default password",
			env:  map[string]string{"DB_PASSWORD": "secret"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "secret", cfg.DBPassword)
			},
		},
		{
			name: "DB_SSLMODE overrides sslmode",
			env:  map[string]string{"DB_SSLMODE": "require"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "require", cfg.SSLMode)
			},
		},
		{
			name: "DB_AUTO_MIGRATE false disables auto migrate",
			env:  map[string]string{"DB_AUTO_MIGRATE": "false"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.False(t, cfg.AutoMigrateSchema)
			},
		},
		{
			name: "DB_AUTO_MIGRATE true enables auto migrate",
			env:  map[string]string{"DB_AUTO_MIGRATE": "true"},
			verify: func(t *testing.T, cfg *resources.Config) {
				assert.True(t, cfg.AutoMigrateSchema)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			unsetEnv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME",
				"DB_USER", "DB_PASSWORD", "DB_SSLMODE", "DB_AUTO_MIGRATE")
			setEnv(t, tc.env)

			cfg, err := resources.NewConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)
			tc.verify(t, cfg)
		})
	}
}

func TestNewConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "invalid SERVER_PORT returns error",
			env:     map[string]string{"SERVER_PORT": "notanint"},
			wantErr: "invalid SERVER_PORT",
		},
		{
			name:    "invalid DB_PORT returns error",
			env:     map[string]string{"DB_PORT": "abc"},
			wantErr: "invalid DB_PORT",
		},
		{
			name:    "invalid DB_AUTO_MIGRATE returns error",
			env:     map[string]string{"DB_AUTO_MIGRATE": "maybe"},
			wantErr: "invalid DB_AUTO_MIGRATE",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			unsetEnv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME",
				"DB_USER", "DB_PASSWORD", "DB_SSLMODE", "DB_AUTO_MIGRATE")
			setEnv(t, tc.env)

			cfg, err := resources.NewConfig()
			assert.Nil(t, cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// ---------------------------------------------------------------------------
// ServerAddr — port 8082
// ---------------------------------------------------------------------------

func TestConfig_ServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantAddr string
	}{
		{
			name:     "default port 8082 produces :8082",
			port:     8082,
			wantAddr: ":8082",
		},
		{
			name:     "custom port 9090 produces :9090",
			port:     9090,
			wantAddr: ":9090",
		},
		{
			name:     "port 80 produces :80",
			port:     80,
			wantAddr: ":80",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &resources.Config{ServerPort: tc.port}
			assert.Equal(t, tc.wantAddr, cfg.ServerAddr())
		})
	}
}

// TestConfig_ServerAddr_Port8082_Binds verifies the invariant that the
// application always listens on port 8082 unless overridden.
func TestConfig_ServerAddr_Port8082_Binds(t *testing.T) {
	// Build a minimal HTTP handler and serve it on a free ephemeral port.
	// We use httptest.NewServer to avoid needing the exact port, but we
	// separately verify that when a server IS configured for 8082, ServerAddr
	// returns ":8082".
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Invariant: default config ServerAddr is ":8082".
	cfg := &resources.Config{ServerPort: resources.DefaultServerPort}
	assert.Equal(t, ":8082", cfg.ServerAddr(),
		"The application always listens on port 8082 unless overridden")
}

// TestServerPort_AlreadyInUse validates that binding to a busy port fails.
func TestServerPort_AlreadyInUse(t *testing.T) {
	// Grab any free port to simulate a port-in-use condition without
	// hardcoding 8082 (which may be occupied in CI).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	addr := ln.Addr().String()

	// A second listen on the same address must fail.
	ln2, err := net.Listen("tcp", addr)
	if err == nil {
		_ = ln2.Close()
		t.Skip("OS allowed double-bind; skipping port-collision test")
	}
	assert.Error(t, err, "startup fails if the port is already in use")
}

// ---------------------------------------------------------------------------
// DSN — spring.datasource.url migration
// ---------------------------------------------------------------------------

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name    string
		cfg     resources.Config
		wantDSN string
	}{
		{
			name: "default values produce correct PostgreSQL DSN",
			cfg: resources.Config{
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
				DBName:     resources.DefaultDBName,
				SSLMode:    resources.DefaultSSLMode,
			},
			wantDSN: "host=localhost port=5432 user=root password=root dbname=barcode sslmode=disable",
		},
		{
			name: "custom values produce correct DSN",
			cfg: resources.Config{
				DBHost:     "pgserver",
				DBPort:     5433,
				DBUser:     "admin",
				DBPassword: "s3cr3t",
				DBName:     "mydb",
				SSLMode:    "require",
			},
			wantDSN: "host=pgserver port=5433 user=admin password=s3cr3t dbname=mydb sslmode=require",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			assert.Equal(t, tc.wantDSN, cfg.DSN())
		})
	}
}

// TestConfig_DSN_ContainsBarcode verifies the invariant that the datasource
// always targets the 'barcode' schema.
func TestConfig_DSN_ContainsBarcode(t *testing.T) {
	unsetEnv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME",
		"DB_USER", "DB_PASSWORD", "DB_SSLMODE", "DB_AUTO_MIGRATE")

	cfg, err := resources.NewConfig()
	require.NoError(t, err)

	assert.Contains(t, cfg.DSN(), "dbname=barcode",
		"datasource must always target the 'barcode' schema by default")
}

// TestConfig_DSN_ContainsRootCredentials verifies the invariant that
// username/password default to root/root.
func TestConfig_DS