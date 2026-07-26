```go
package resources

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setEnv sets a collection of environment variables for the duration of a
// test and restores the previous values via t.Cleanup.
func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		prev, existed := os.LookupEnv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv %s: %v", k, err)
		}
		k, prev, existed := k, prev, existed // capture
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(k, prev)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// unsetEnv unsets a collection of environment variables for the duration of a
// test and restores them via t.Cleanup.
func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		prev, existed := os.LookupEnv(k)
		_ = os.Unsetenv(k)
		k, prev, existed := k, prev, existed
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(k, prev)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Constants / default values
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	tests := []struct {
		name  string
		got   interface{}
		want  interface{}
	}{
		{"DefaultServerPort is 8082", DefaultServerPort, 8082},
		{"DefaultDBHost is localhost", DefaultDBHost, "localhost"},
		{"DefaultDBPort is 5432", DefaultDBPort, 5432},
		{"DefaultDBName is barcode", DefaultDBName, "barcode"},
		{"DefaultDBUser is postgres", DefaultDBUser, "postgres"},
		{"DefaultDBPassword is root", DefaultDBPassword, "root"},
		{"DefaultDBSSLMode is disable", DefaultDBSSLMode, "disable"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got)
		})
	}
}

// ---------------------------------------------------------------------------
// envString
// ---------------------------------------------------------------------------

func TestEnvString(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		envVal  string
		envSet  bool
		def     string
		want    string
	}{
		{
			name:   "returns env value when set and non-empty",
			key:    "TEST_ENV_STR_1",
			envVal: "hello",
			envSet: true,
			def:    "default",
			want:   "hello",
		},
		{
			name:   "returns default when env not set",
			key:    "TEST_ENV_STR_2",
			envSet: false,
			def:    "fallback",
			want:   "fallback",
		},
		{
			name:   "returns default when env is empty string",
			key:    "TEST_ENV_STR_3",
			envVal: "",
			envSet: true,
			def:    "fallback",
			want:   "fallback",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				setEnv(t, map[string]string{tc.key: tc.envVal})
			} else {
				unsetEnv(t, tc.key)
			}
			got := envString(tc.key, tc.def)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// envInt
// ---------------------------------------------------------------------------

func TestEnvInt(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		envVal  string
		envSet  bool
		def     int
		want    int
		wantErr bool
	}{
		{
			name:   "returns env value when set to valid int",
			key:    "TEST_ENV_INT_1",
			envVal: "9090",
			envSet: true,
			def:    8080,
			want:   9090,
		},
		{
			name:   "returns default when env not set",
			key:    "TEST_ENV_INT_2",
			envSet: false,
			def:    8080,
			want:   8080,
		},
		{
			name:   "returns default when env is empty",
			key:    "TEST_ENV_INT_3",
			envVal: "",
			envSet: true,
			def:    8080,
			want:   8080,
		},
		{
			name:    "returns error when env is non-numeric",
			key:     "TEST_ENV_INT_4",
			envVal:  "not-a-number",
			envSet:  true,
			def:     8080,
			want:    0,
			wantErr: true,
		},
		{
			name:    "returns error when env is float-like",
			key:     "TEST_ENV_INT_5",
			envVal:  "3.14",
			envSet:  true,
			def:     8080,
			want:    0,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				setEnv(t, map[string]string{tc.key: tc.envVal})
			} else {
				unsetEnv(t, tc.key)
			}
			got, err := envInt(tc.key, tc.def)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig
// ---------------------------------------------------------------------------

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		unset   []string
		want    *Config
		wantErr bool
		errContains string
	}{
		{
			name: "all defaults when no env vars set",
			unset: []string{
				"SERVER_PORT", "DB_HOST", "DB_PORT",
				"DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
			},
			want: &Config{
				ServerPort: DefaultServerPort,
				DBHost:     DefaultDBHost,
				DBPort:     DefaultDBPort,
				DBName:     DefaultDBName,
				DBUser:     DefaultDBUser,
				DBPassword: DefaultDBPassword,
				DBSSLMode:  DefaultDBSSLMode,
			},
		},
		{
			name: "all values from environment",
			env: map[string]string{
				"SERVER_PORT": "9000",
				"DB_HOST":     "db.example.com",
				"DB_PORT":     "5433",
				"DB_NAME":     "mydb",
				"DB_USER":     "alice",
				"DB_PASSWORD": "s3cr3t",
				"DB_SSLMODE":  "require",
			},
			want: &Config{
				ServerPort: 9000,
				DBHost:     "db.example.com",
				DBPort:     5433,
				DBName:     "mydb",
				DBUser:     "alice",
				DBPassword: "s3cr3t",
				DBSSLMode:  "require",
			},
		},
		{
			name: "migrated defaults match original intent – server port 8082",
			unset: []string{"SERVER_PORT"},
			want: &Config{
				ServerPort: 8082,
				DBHost:     DefaultDBHost,
				DBPort:     DefaultDBPort,
				DBName:     DefaultDBName,
				DBUser:     DefaultDBUser,
				DBPassword: DefaultDBPassword,
				DBSSLMode:  DefaultDBSSLMode,
			},
		},
		{
			name: "migrated defaults match original intent – db name barcode",
			unset: []string{"DB_NAME"},
			want: &Config{
				ServerPort: DefaultServerPort,
				DBHost:     DefaultDBHost,
				DBPort:     DefaultDBPort,
				DBName:     "barcode",
				DBUser:     DefaultDBUser,
				DBPassword: DefaultDBPassword,
				DBSSLMode:  DefaultDBSSLMode,
			},
		},
		{
			name: "migrated defaults – password is root (mirrors spring.datasource.password)",
			unset: []string{"DB_PASSWORD"},
			want: &Config{
				ServerPort: DefaultServerPort,
				DBHost:     DefaultDBHost,
				DBPort:     DefaultDBPort,
				DBName:     DefaultDBName,
				DBUser:     DefaultDBUser,
				DBPassword: "root",
				DBSSLMode:  DefaultDBSSLMode,
			},
		},
		{
			name: "invalid SERVER_PORT returns error",
			env: map[string]string{
				"SERVER_PORT": "bad-port",
			},
			wantErr:     true,
			errContains: "SERVER_PORT",
		},
		{
			name: "invalid DB_PORT returns error",
			env: map[string]string{
				"SERVER_PORT": "8082",
				"DB_PORT":     "not-a-number",
			},
			wantErr:     true,
			errContains: "DB_PORT",
		},
		{
			name: "partial override – only DB_HOST set",
			env: map[string]string{
				"DB_HOST": "remote-host",
			},
			unset: []string{
				"SERVER_PORT", "DB_PORT", "DB_NAME",
				"DB_USER", "DB_PASSWORD", "DB_SSLMODE",
			},
			want: &Config{
				ServerPort: DefaultServerPort,
				DBHost:     "remote-host",
				DBPort:     DefaultDBPort,
				DBName:     DefaultDBName,
				DBUser:     DefaultDBUser,
				DBPassword: DefaultDBPassword,
				DBSSLMode:  DefaultDBSSLMode,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Apply environment mutations only for the duration of this sub-test.
			if len(tc.env) > 0 {
				setEnv(t, tc.env)
			}
			if len(tc.unset) > 0 {
				unsetEnv(t, tc.unset...)
			}

			got, err := LoadConfig()

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Config.DSN
// ---------------------------------------------------------------------------

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default config produces expected DSN",
			cfg: Config{
				DBHost:     DefaultDBHost,
				DBPort:     DefaultDBPort,
				DBUser:     DefaultDBUser,
				DBPassword: DefaultDBPassword,
				DBName:     DefaultDBName,
				DBSSLMode:  DefaultDBSSLMode,
			},
			want: fmt.Sprintf(
				"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				DefaultDBHost, DefaultDBPort, DefaultDBUser,
				DefaultDBPassword, DefaultDBName, DefaultDBSSLMode,
			),
		},
		{
			name: "custom values are reflected in DSN",
			cfg: Config{
				DBHost:     "db.prod.example.com",
				DBPort:     5433,
				DBUser:     "alice",
				DBPassword: "hunter2",
				DBName:     "myapp",
				DBSSLMode:  "require",
			},
			want: "host=db.prod.example.com port=5433 user=alice password=hunter2 dbname=myapp sslmode=require",
		},
		{
			name: "DSN is PostgreSQL-shaped (not JDBC MySQL URL)",
			cfg: Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "root",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			// Must NOT contain jdbc:mysql style artefacts.
			want: "host=localhost port=5432 user=postgres password=root dbname=barcode sslmode=disable",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.DSN()
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestConfig_DSN_NoJDBC verifies that the migrated DSN does not contain MySQL
// JDBC artefacts (requirement from the dialect-switch migration note).
func TestConfig_DSN_NoJDBC(t *testing.T) {
	unsetEnv(t,
		"SERVER_PORT", "DB_HOST", "DB_PORT",
		"DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
	)
	cfg, err := LoadConfig()
	require.NoError(t, err)

	dsn := cfg.DSN()
	assert.NotContains(t, dsn, "jdbc:", "DSN must not contain JDBC prefix")
	assert.NotContains(t, dsn, "mysql", "DSN must not reference MySQL")
	assert.NotContains(t, dsn, "3306", "DSN must not use MySQL default port")
}

// ---------------------------------------------------------------------------
// Config.ServerAddr
// ---------------------------------------------------------------------------

func TestConfig_ServerAddr(t *testing.T) {
	tests := []struct {
		name       string
		serverPort int
		want       string
	}{
		{"default port 8082 produces :8082", 8082, ":8082"},
		{"custom port 9000 produces :9000", 9000, ":9000"},
		{"port 80 produces :80", 80, ":80"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{ServerPort: tc.serverPort}
			assert.Equal(t, tc.want, cfg.ServerAddr())
		})
	}
}

// ---------------------------------------------------------------------------
// Config.OpenDB – uses a mock/stub; avoids real network calls
// ---------------------------------------------------------------------------

// TestConfig_OpenDB_InvalidDriver verifies that OpenDB returns a meaningful
// error when the driver name is not registered. We rely on the fact that
// "postgres" is not registered in this test binary (no driver blank-import).
func TestConfig_OpenDB_InvalidDriver(t *testing.T) {
	// If a Postgres driver happens to be registered in this test binary and a
	// real DB is reachable, we skip rather than fail — the focus is on the
	// no-DB error path.
	cfg := &Config{
		DBHost:     "