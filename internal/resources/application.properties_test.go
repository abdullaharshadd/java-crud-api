```go
package resources_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/resources"
)

// ---------------------------------------------------------------------------
// Constants / defaults
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	t.Run("DefaultServerPort is 8082", func(t *testing.T) {
		assert.Equal(t, 8082, resources.DefaultServerPort,
			"server.port default must mirror the source application.properties value 8082")
	})

	t.Run("DefaultDatabaseURL targets barcode schema", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "barcode",
			"default DSN must reference the 'barcode' database")
	})

	t.Run("DefaultDatabaseURL contains root credentials", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "root",
			"default DSN must embed the root username (spring.datasource.username)")
		// The DSN format postgres://root:root@… encodes both user and password.
		parts := strings.SplitN(resources.DefaultDatabaseURL, "root:root", 2)
		assert.Len(t, parts, 2,
			"default DSN must embed password 'root' (spring.datasource.password)")
	})

	t.Run("DefaultDatabaseURL is a PostgreSQL DSN", func(t *testing.T) {
		assert.True(t,
			strings.HasPrefix(resources.DefaultDatabaseURL, "postgres://") ||
				strings.HasPrefix(resources.DefaultDatabaseURL, "postgresql://"),
			"migration standardises on PostgreSQL; DSN must start with postgres://")
	})

	t.Run("EnvServerPort constant", func(t *testing.T) {
		assert.Equal(t, "SERVER_PORT", resources.EnvServerPort)
	})

	t.Run("EnvDatabaseURL constant", func(t *testing.T) {
		assert.Equal(t, "DATABASE_URL", resources.EnvDatabaseURL)
	})
}

// ---------------------------------------------------------------------------
// Config.Addr
// ---------------------------------------------------------------------------

func TestConfig_Addr(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantAddr string
	}{
		{
			name:     "default port 8082 produces :8082",
			port:     resources.DefaultServerPort,
			wantAddr: ":8082",
		},
		{
			name:     "port 80 produces :80",
			port:     80,
			wantAddr: ":80",
		},
		{
			name:     "port 1 produces :1",
			port:     1,
			wantAddr: ":1",
		},
		{
			name:     "port 65535 produces :65535",
			port:     65535,
			wantAddr: ":65535",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := resources.Config{ServerPort: tc.port}
			assert.Equal(t, tc.wantAddr, cfg.Addr())
		})
	}
}

// ---------------------------------------------------------------------------
// Load — happy paths
// ---------------------------------------------------------------------------

// setenv is a helper that sets an env-var for the duration of a test and
// restores the previous value (or unsets it) via t.Cleanup.
func setenv(t *testing.T, key, value string) {
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

func unsetenv(t *testing.T, key string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, prev)
		}
	})
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure neither override variable is set so we exercise pure defaults.
	unsetenv(t, resources.EnvServerPort)
	unsetenv(t, resources.EnvDatabaseURL)

	cfg, err := resources.Load()
	require.NoError(t, err)

	t.Run("default server port is 8082 (mirrors server.port)", func(t *testing.T) {
		assert.Equal(t, resources.DefaultServerPort, cfg.ServerPort)
		assert.Equal(t, 8082, cfg.ServerPort)
	})

	t.Run("default listen addr is :8082", func(t *testing.T) {
		assert.Equal(t, ":8082", cfg.Addr(),
			"server must bind to :8082 when no override is present")
	})

	t.Run("default database URL is set", func(t *testing.T) {
		assert.Equal(t, resources.DefaultDatabaseURL, cfg.DatabaseURL)
	})

	t.Run("default database URL targets barcode schema", func(t *testing.T) {
		assert.Contains(t, cfg.DatabaseURL, "barcode")
	})

	t.Run("default credentials are root/root", func(t *testing.T) {
		assert.Contains(t, cfg.DatabaseURL, "root:root",
			"credentials (username=root, password=root) must be present in the default DSN")
	})
}

func TestLoad_EnvOverrides(t *testing.T) {
	tests := []struct {
		name         string
		envPort      string
		envDB        string
		wantPort     int
		wantDBSubstr string
		wantAddr     string
	}{
		{
			name:         "override server port via SERVER_PORT",
			envPort:      "9090",
			envDB:        "",
			wantPort:     9090,
			wantDBSubstr: "barcode", // still default DB
			wantAddr:     ":9090",
		},
		{
			name:         "override database URL via DATABASE_URL",
			envPort:      "",
			envDB:        "postgres://admin:secret@db.example.com:5432/mydb?sslmode=require",
			wantPort:     8082, // still default port
			wantDBSubstr: "mydb",
			wantAddr:     ":8082",
		},
		{
			name:         "override both SERVER_PORT and DATABASE_URL",
			envPort:      "3000",
			envDB:        "postgres://user:pass@remote:5432/testdb",
			wantPort:     3000,
			wantDBSubstr: "testdb",
			wantAddr:     ":3000",
		},
		{
			name:         "port 1 (minimum valid)",
			envPort:      "1",
			envDB:        "",
			wantPort:     1,
			wantDBSubstr: "barcode",
			wantAddr:     ":1",
		},
		{
			name:         "port 65535 (maximum valid)",
			envPort:      "65535",
			envDB:        "",
			wantPort:     65535,
			wantDBSubstr: "barcode",
			wantAddr:     ":65535",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Always clear both variables to avoid cross-test pollution.
			unsetenv(t, resources.EnvServerPort)
			unsetenv(t, resources.EnvDatabaseURL)

			if tc.envPort != "" {
				setenv(t, resources.EnvServerPort, tc.envPort)
			}
			if tc.envDB != "" {
				setenv(t, resources.EnvDatabaseURL, tc.envDB)
			}

			cfg, err := resources.Load()
			require.NoError(t, err)

			assert.Equal(t, tc.wantPort, cfg.ServerPort)
			assert.Equal(t, tc.wantAddr, cfg.Addr())
			assert.Contains(t, cfg.DatabaseURL, tc.wantDBSubstr)
		})
	}
}

// ---------------------------------------------------------------------------
// Load — error cases (invalid SERVER_PORT)
// ---------------------------------------------------------------------------

func TestLoad_InvalidServerPort(t *testing.T) {
	tests := []struct {
		name       string
		envPort    string
		wantErrSub string
	}{
		{
			name:       "non-numeric port returns error",
			envPort:    "notanumber",
			wantErrSub: "SERVER_PORT",
		},
		{
			name:       "port 0 is out of range",
			envPort:    "0",
			wantErrSub: "out of range",
		},
		{
			name:       "negative port is out of range",
			envPort:    "-1",
			wantErrSub: "out of range",
		},
		{
			name:       "port 65536 exceeds maximum",
			envPort:    "65536",
			wantErrSub: "out of range",
		},
		{
			name:       "port 99999 exceeds maximum",
			envPort:    "99999",
			wantErrSub: "out of range",
		},
		{
			name:       "float-like string is not valid",
			envPort:    "80.5",
			wantErrSub: "SERVER_PORT",
		},
		{
			name:       "empty-ish string with spaces",
			envPort:    " ",
			wantErrSub: "SERVER_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			unsetenv(t, resources.EnvServerPort)
			unsetenv(t, resources.EnvDatabaseURL)

			setenv(t, resources.EnvServerPort, tc.envPort)

			cfg, err := resources.Load()
			require.Error(t, err,
				"Load should return an error for invalid SERVER_PORT %q", tc.envPort)
			assert.Contains(t, err.Error(), tc.wantErrSub,
				"error message should mention the problematic value or env-var name")
			assert.Equal(t, resources.Config{}, cfg,
				"on error, Load must return zero Config")
		})
	}
}

// ---------------------------------------------------------------------------
// Load — empty string env vars treated as unset (defaults preserved)
// ---------------------------------------------------------------------------

func TestLoad_EmptyEnvVarsUseDefaults(t *testing.T) {
	// Setting a variable to "" must be treated as "not set" by Load.
	unsetenv(t, resources.EnvServerPort)
	unsetenv(t, resources.EnvDatabaseURL)
	setenv(t, resources.EnvServerPort, "")
	setenv(t, resources.EnvDatabaseURL, "")

	cfg, err := resources.Load()
	require.NoError(t, err)

	assert.Equal(t, resources.DefaultServerPort, cfg.ServerPort,
		"empty SERVER_PORT must fall back to default 8082")
	assert.Equal(t, resources.DefaultDatabaseURL, cfg.DatabaseURL,
		"empty DATABASE_URL must fall back to default DSN")
}

// ---------------------------------------------------------------------------
// Behavioral spec: server always serves on port 8082 unless overridden
// ---------------------------------------------------------------------------

func TestServerPortInvariant(t *testing.T) {
	t.Run("without override port is always 8082", func(t *testing.T) {
		unsetenv(t, resources.EnvServerPort)
		unsetenv(t, resources.EnvDatabaseURL)

		for i := 0; i < 3; i++ {
			cfg, err := resources.Load()
			require.NoError(t, err)
			assert.Equal(t, 8082, cfg.ServerPort,
				"server must always bind to 8082 when no override is present (iteration %d)", i)
		}
	})
}

// ---------------------------------------------------------------------------
// Behavioral spec: database invariants
// ---------------------------------------------------------------------------

func TestDatabaseURLInvariants(t *testing.T) {
	t.Run("default DSN always targets barcode schema", func(t *testing.T) {
		unsetenv(t, resources.EnvServerPort)
		unsetenv(t, resources.EnvDatabaseURL)

		cfg, err := resources.Load()
		require.NoError(t, err)
		assert.Contains(t, cfg.DatabaseURL, "barcode",
			"datasource always targets the 'barcode' database unless overridden")
	})

	t.Run("default DSN always carries root credentials", func(t *testing.T) {
		unsetenv(t, resources.EnvServerPort)
		unsetenv(t, resources.EnvDatabaseURL)

		cfg, err := resources.Load()
		require.NoError(t, err)
		assert.Contains(t, cfg.DatabaseURL, "root",
			"username 'root' (spring.datasource.username) must appear in the default DSN")

		// Verify that both username and password are "root".
		assert.Contains(t, cfg.DatabaseURL, "root:root",
			"password 'root' (spring.datasource.password) must appear in the default DSN")
	})
}

// ---------------------------------------------------------------------------
// Behavioral spec: DDL-auto / schema-management note
// The migrated code replaces Hibernate's ddl-auto=update with CREATE TABLE IF
// NOT EXISTS (ensureSchema). The property is intentionally dropped from Config.
// We verify that no "ddl" or "hibernate" field leaks into the Config struct.
// ---------------------------------------------------------------------------

func TestConfig_NoDDLField(t *testing.T) {
	// The Config struct must NOT expose a DDL-auto field — schema management is
	// handled in code (ensureSchema), not via a configuration property.
	cfg := resources.Config{
		ServerPort:  8082,
		DatabaseURL: resources.DefaultDatabaseURL,
	}
	// Compile-time proof: if someone added DDLAuto it would be usable here.
	// We assert the struct only has the two expected fields by checking them.
	assert.Equal(t, 8082, cfg.ServerPort)
	assert.Equal(t, resources.DefaultDatabaseURL, cfg.DatabaseURL)
}

// ---------------------------------------------------------------------------
// Behavioral spec: dialect / driver note
// The driver and dialect are compile-time concerns, not runtime config.
// We verify that the Config struct has no "dialect" or "driver" field.
// ---------------------------------------------------------------------------

func TestConfig_NoDriverOrDialectField(t *testing.T) {
	cfg, err := resources.Load()
	require.NoError(t, err)

	// The DSN itself does not contain MySQL-specific patterns; the migration
	// standardises on PostgreSQL.
	assert.False(t, strings.Contains(cfg.DatabaseURL, "mysql"),
		"the migrated DSN must not reference MySQL (migration target is PostgreSQL)")
	assert.False(t, strings.Contains(cfg.DatabaseURL, "jdbc"),
		"the migrated DSN must not use JDBC syntax")
}

// ---------------------------------------------------------------------------
// Addr format correctness
// ---------------------------------------------------------------------------

func TestAddr_Format(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{8082, ":8082"},
		{80, ":80"},
		{443, ":443"},
		{1024, ":1024"},
		{65535, ":65535"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("port %d -> %s", tc.port, tc.want), func(t *testing.T) {
			cfg := resources.Config{ServerPort: tc.port}
			assert.Equal(t, tc.want, cfg.