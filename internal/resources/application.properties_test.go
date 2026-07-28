```go
package resources_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"smartcontact/internal/resources"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// setEnv sets a collection of environment variables and returns a cleanup
// function that restores the previous state.
func setEnv(t *testing.T, pairs map[string]string) func() {
	t.Helper()
	prev := make(map[string]string, len(pairs))
	wasSet := make(map[string]bool, len(pairs))
	for k := range pairs {
		v, ok := os.LookupEnv(k)
		prev[k] = v
		wasSet[k] = ok
	}
	for k, v := range pairs {
		require.NoError(t, os.Setenv(k, v))
	}
	return func() {
		for k := range pairs {
			if wasSet[k] {
				_ = os.Setenv(k, prev[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

// unsetEnv temporarily removes a list of environment variables.
func unsetEnv(t *testing.T, keys ...string) func() {
	t.Helper()
	prev := make(map[string]string, len(keys))
	wasSet := make(map[string]bool, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		prev[k] = v
		wasSet[k] = ok
		_ = os.Unsetenv(k)
	}
	return func() {
		for _, k := range keys {
			if wasSet[k] {
				_ = os.Setenv(k, prev[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – defaults
// ---------------------------------------------------------------------------

func TestLoadConfig_Defaults(t *testing.T) {
	allKeys := []string{
		"SERVER_PORT", "DB_HOST", "DB_PORT",
		"DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
	}
	cleanup := unsetEnv(t, allKeys...)
	defer cleanup()

	cfg, err := resources.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"default ServerPort equals 8082 (server.port)", cfg.ServerPort, 8082},
		{"default DBHost equals localhost", cfg.DBHost, "localhost"},
		{"default DBPort equals 5432 (PostgreSQL)", cfg.DBPort, 5432},
		{"default DBName equals barcode (datasource.url db segment)", cfg.DBName, "barcode"},
		// The migration replaces root/root MySQL creds with postgres/postgres
		// as local-dev fallbacks; the *intent* of the original property is
		// preserved (a default credential that must be overridden in prod).
		{"default DBUser carries a default credential (datasource.username intent)", cfg.DBUser, "postgres"},
		{"default DBPassword carries a default credential (datasource.password intent)", cfg.DBPassword, "postgres"},
		{"default DBSSLMode is disable", cfg.DBSSLMode, "disable"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – SERVER_PORT override
// ---------------------------------------------------------------------------

func TestLoadConfig_ServerPort(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantPort    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid numeric SERVER_PORT is accepted",
			envValue: "9090",
			wantPort: 9090,
		},
		{
			name:     "port 8082 explicitly set equals default",
			envValue: "8082",
			wantPort: 8082,
		},
		{
			name:        "non-numeric SERVER_PORT returns error",
			envValue:    "not-a-port",
			wantErr:     true,
			errContains: "SERVER_PORT",
		},
		{
			name:        "empty string SERVER_PORT returns error",
			envValue:    "",
			wantErr:     true,
			errContains: "SERVER_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"SERVER_PORT": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPort, cfg.ServerPort)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_HOST override
// ---------------------------------------------------------------------------

func TestLoadConfig_DBHost(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantHost string
	}{
		{"override with custom host", "db.example.com", "db.example.com"},
		{"override with IP address", "192.168.1.50", "192.168.1.50"},
		{"override with empty string clears host", "", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"DB_HOST": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantHost, cfg.DBHost)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_PORT override
// ---------------------------------------------------------------------------

func TestLoadConfig_DBPort(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantPort    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid PostgreSQL port 5432",
			envValue: "5432",
			wantPort: 5432,
		},
		{
			name:     "custom port 5433 accepted",
			envValue: "5433",
			wantPort: 5433,
		},
		{
			name:        "non-numeric DB_PORT returns error",
			envValue:    "abc",
			wantErr:     true,
			errContains: "DB_PORT",
		},
		{
			name:        "empty DB_PORT returns error",
			envValue:    "",
			wantErr:     true,
			errContains: "DB_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"DB_PORT": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPort, cfg.DBPort)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_NAME override
// (mirrors spring.datasource.url database segment = "barcode")
// ---------------------------------------------------------------------------

func TestLoadConfig_DBName(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantName string
	}{
		{"override db name", "mydb", "mydb"},
		// The original source always targeted 'barcode'; ensure the default
		// is preserved when the env var is absent.
		{"absent env retains barcode default", "", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue == "" && tc.wantName == "" {
				// Test the absent case without setting the env var.
				cleanup := unsetEnv(t, "DB_NAME")
				defer cleanup()
				cfg, err := resources.LoadConfig()
				require.NoError(t, err)
				assert.Equal(t, "barcode", cfg.DBName,
					"default DBName must be 'barcode' when DB_NAME is not set")
				return
			}
			cleanup := setEnv(t, map[string]string{"DB_NAME": tc.envValue})
			defer cleanup()
			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.envValue, cfg.DBName)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_USER override
// (mirrors spring.datasource.username)
// ---------------------------------------------------------------------------

func TestLoadConfig_DBUser(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantUser string
	}{
		{"custom user overrides default", "appuser", "appuser"},
		{"admin user accepted", "admin", "admin"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"DB_USER": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, cfg.DBUser)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_PASSWORD override
// (mirrors spring.datasource.password)
// ---------------------------------------------------------------------------

func TestLoadConfig_DBPassword(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantPass string
	}{
		{"custom password overrides default", "s3cr3t!", "s3cr3t!"},
		{"empty password accepted (cleared)", "", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"DB_PASSWORD": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantPass, cfg.DBPassword)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – DB_SSLMODE override
// ---------------------------------------------------------------------------

func TestLoadConfig_DBSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantSSLMode string
	}{
		{"require mode for production", "require", "require"},
		{"verify-full mode", "verify-full", "verify-full"},
		{"disable mode explicit", "disable", "disable"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{"DB_SSLMODE": tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantSSLMode, cfg.DBSSLMode)
		})
	}
}

// ---------------------------------------------------------------------------
// LoadConfig – multiple overrides together
// ---------------------------------------------------------------------------

func TestLoadConfig_MultipleEnvOverrides(t *testing.T) {
	envs := map[string]string{
		"SERVER_PORT": "3000",
		"DB_HOST":     "pghost",
		"DB_PORT":     "5433",
		"DB_NAME":     "testdb",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "testpass",
		"DB_SSLMODE":  "require",
	}
	cleanup := setEnv(t, envs)
	defer cleanup()

	cfg, err := resources.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 3000, cfg.ServerPort)
	assert.Equal(t, "pghost", cfg.DBHost)
	assert.Equal(t, 5433, cfg.DBPort)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testpass", cfg.DBPassword)
	assert.Equal(t, "require", cfg.DBSSLMode)
}

// ---------------------------------------------------------------------------
// LoadConfig – error cases return nil config
// ---------------------------------------------------------------------------

func TestLoadConfig_ErrorCasesReturnNilConfig(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envValue    string
		errContains string
	}{
		{
			name:        "invalid SERVER_PORT returns nil config",
			envKey:      "SERVER_PORT",
			envValue:    "bad",
			errContains: "SERVER_PORT",
		},
		{
			name:        "invalid DB_PORT returns nil config",
			envKey:      "DB_PORT",
			envValue:    "bad",
			errContains: "DB_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setEnv(t, map[string]string{tc.envKey: tc.envValue})
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.Error(t, err)
			assert.Nil(t, cfg, "LoadConfig must return nil when an error occurs")
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// ---------------------------------------------------------------------------
// Config.DSN
// ---------------------------------------------------------------------------

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name       string
		cfg        resources.Config
		wantFields []string // substrings that must appear in the DSN
		notWant    []string // substrings that must NOT appear (e.g. MySQL artefacts)
	}{
		{
			name: "DSN contains all PostgreSQL key=value pairs",
			cfg: resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			wantFields: []string{
				"host=localhost",
				"port=5432",
				"user=postgres",
				"password=postgres",
				"dbname=barcode",
				"sslmode=disable",
			},
			// The migration explicitly forbids MySQL artefacts.
			notWant: []string{"jdbc:", "mysql", "3306"},
		},
		{
			name: "DSN reflects custom host and port",
			cfg: resources.Config{
				DBHost:     "pghost.example.com",
				DBPort:     5433,
				DBUser:     "appuser",
				DBPassword: "s3cr3t",
				DBName:     "myapp",
				DBSSLMode:  "require",
			},
			wantFields: []string{
				"host=pghost.example.com",
				"port=5433",
				"user=appuser",
				"password=s