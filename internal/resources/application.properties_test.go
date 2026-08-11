```go
package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/resources"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// setEnv sets a map of environment variables and returns a cleanup function
// that restores the original values (or unsets them if they were not present).
func setEnv(t *testing.T, vars map[string]string) func() {
	t.Helper()
	original := make(map[string]string, len(vars))
	wasSet := make(map[string]bool, len(vars))

	for k, v := range vars {
		old, ok := os.LookupEnv(k)
		original[k] = old
		wasSet[k] = ok
		require.NoError(t, os.Setenv(k, v))
	}

	return func() {
		for k := range vars {
			if wasSet[k] {
				_ = os.Setenv(k, original[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

// unsetEnv temporarily unsets the listed environment variables and returns a
// cleanup function that restores them.
func unsetEnv(t *testing.T, keys ...string) func() {
	t.Helper()
	original := make(map[string]string, len(keys))
	wasSet := make(map[string]bool, len(keys))

	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		original[k] = v
		wasSet[k] = ok
		_ = os.Unsetenv(k)
	}

	return func() {
		for _, k := range keys {
			if wasSet[k] {
				_ = os.Setenv(k, original[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

// allEnvKeys lists every environment variable that LoadConfig reads so that
// tests can guarantee a clean slate.
var allEnvKeys = []string{
	"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME",
	"DB_USER", "DB_PASSWORD", "DB_SSLMODE",
}

// ---------------------------------------------------------------------------
// TestLoadConfig_Defaults
// ---------------------------------------------------------------------------

// TestLoadConfig_Defaults verifies that LoadConfig returns the correct
// compile-time defaults when no environment variables are set.
// Covers: spec "Embedded server port configuration" invariant (port 8082)
//         and the migration note that development defaults mirror the source
//         application.properties.
func TestLoadConfig_Defaults(t *testing.T) {
	restore := unsetEnv(t, allEnvKeys...)
	defer restore()

	cfg, err := resources.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"ServerPort default is 8082", cfg.ServerPort, 8082},
		{"DBHost default is localhost", cfg.DBHost, "localhost"},
		{"DBPort default is 5432 (PostgreSQL)", cfg.DBPort, 5432},
		{"DBName default is barcode", cfg.DBName, "barcode"},
		{"DBUser default is postgres", cfg.DBUser, "postgres"},
		{"DBPassword default is postgres", cfg.DBPassword, "postgres"},
		{"DBSSLMode default is disable", cfg.DBSSLMode, "disable"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestLoadConfig_EnvOverrides
// ---------------------------------------------------------------------------

// TestLoadConfig_EnvOverrides verifies that every recognized environment
// variable takes precedence over the compiled-in default.
// Covers: global invariant "all configuration values … may be overridden by
//         environment variables".
func TestLoadConfig_EnvOverrides(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(t *testing.T, cfg *resources.Config)
	}{
		{
			name:    "SERVER_PORT overrides default port",
			envVars: map[string]string{"SERVER_PORT": "9090"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 9090, cfg.ServerPort)
			},
		},
		{
			name:    "DB_HOST overrides default host",
			envVars: map[string]string{"DB_HOST": "db.example.com"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "db.example.com", cfg.DBHost)
			},
		},
		{
			name:    "DB_PORT overrides default port",
			envVars: map[string]string{"DB_PORT": "5433"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 5433, cfg.DBPort)
			},
		},
		{
			name:    "DB_NAME overrides default name",
			envVars: map[string]string{"DB_NAME": "production_db"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "production_db", cfg.DBName)
			},
		},
		{
			name:    "DB_USER overrides default user",
			envVars: map[string]string{"DB_USER": "admin"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "admin", cfg.DBUser)
			},
		},
		{
			name:    "DB_PASSWORD overrides default password",
			envVars: map[string]string{"DB_PASSWORD": "s3cr3t"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "s3cr3t", cfg.DBPassword)
			},
		},
		{
			name:    "DB_SSLMODE overrides default sslmode",
			envVars: map[string]string{"DB_SSLMODE": "require"},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "require", cfg.DBSSLMode)
			},
		},
		{
			name: "all env vars overridden simultaneously",
			envVars: map[string]string{
				"SERVER_PORT": "7070",
				"DB_HOST":     "remote.db",
				"DB_PORT":     "5434",
				"DB_NAME":     "mydb",
				"DB_USER":     "myuser",
				"DB_PASSWORD": "mypass",
				"DB_SSLMODE":  "verify-full",
			},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 7070, cfg.ServerPort)
				assert.Equal(t, "remote.db", cfg.DBHost)
				assert.Equal(t, 5434, cfg.DBPort)
				assert.Equal(t, "mydb", cfg.DBName)
				assert.Equal(t, "myuser", cfg.DBUser)
				assert.Equal(t, "mypass", cfg.DBPassword)
				assert.Equal(t, "verify-full", cfg.DBSSLMode)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			restore := unsetEnv(t, allEnvKeys...)
			defer restore()

			cleanup := setEnv(t, tt.envVars)
			defer cleanup()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)
			tt.check(t, cfg)
		})
	}
}

// ---------------------------------------------------------------------------
// TestLoadConfig_InvalidEnvReturnsError
// ---------------------------------------------------------------------------

// TestLoadConfig_InvalidEnvReturnsError verifies that non-integer values for
// numeric env-vars cause LoadConfig to return a descriptive error.
// Covers: error cases in intFromEnv.
func TestLoadConfig_InvalidEnvReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		errContains string
	}{
		{
			name:        "invalid SERVER_PORT returns error",
			envVars:     map[string]string{"SERVER_PORT": "not-a-number"},
			errContains: "SERVER_PORT",
		},
		{
			name:        "float SERVER_PORT returns error",
			envVars:     map[string]string{"SERVER_PORT": "80.5"},
			errContains: "SERVER_PORT",
		},
		{
			name:        "invalid DB_PORT returns error",
			envVars:     map[string]string{"DB_PORT": "abc"},
			errContains: "DB_PORT",
		},
		{
			name:        "DB_PORT with spaces returns error",
			envVars:     map[string]string{"DB_PORT": "54 32"},
			errContains: "DB_PORT",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			restore := unsetEnv(t, allEnvKeys...)
			defer restore()

			cleanup := setEnv(t, tt.envVars)
			defer cleanup()

			cfg, err := resources.LoadConfig()
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// ---------------------------------------------------------------------------
// TestLoadConfig_EmptyEnvFallsBackToDefault
// ---------------------------------------------------------------------------

// TestLoadConfig_EmptyEnvFallsBackToDefault verifies that an empty string for
// a numeric env-var is treated the same as unset (falls back to the default)
// rather than returning an error.
func TestLoadConfig_EmptyEnvFallsBackToDefault(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(t *testing.T, cfg *resources.Config)
	}{
		{
			name:    "empty SERVER_PORT falls back to 8082",
			envVars: map[string]string{"SERVER_PORT": ""},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 8082, cfg.ServerPort)
			},
		},
		{
			name:    "empty DB_PORT falls back to 5432",
			envVars: map[string]string{"DB_PORT": ""},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, 5432, cfg.DBPort)
			},
		},
		{
			name:    "empty DB_HOST falls back to localhost",
			envVars: map[string]string{"DB_HOST": ""},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "localhost", cfg.DBHost)
			},
		},
		{
			name:    "empty DB_USER falls back to postgres",
			envVars: map[string]string{"DB_USER": ""},
			check: func(t *testing.T, cfg *resources.Config) {
				assert.Equal(t, "postgres", cfg.DBUser)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			restore := unsetEnv(t, allEnvKeys...)
			defer restore()

			// Explicitly set variables to empty string (different from unset).
			for k, v := range tt.envVars {
				require.NoError(t, os.Setenv(k, v))
			}
			defer func() {
				for k := range tt.envVars {
					_ = os.Unsetenv(k)
				}
			}()

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)
			tt.check(t, cfg)
		})
	}
}

// ---------------------------------------------------------------------------
// TestConfig_DSN
// ---------------------------------------------------------------------------

// TestConfig_DSN verifies that Config.DSN() produces a correctly formatted
// lib/pq-style DSN.
// Covers: spec "MySQL datasource configuration" — the Go migration targets
//         PostgreSQL and must emit a valid lib/pq DSN.
func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name   string
		cfg    resources.Config
		wantDSN string
	}{
		{
			name: "default values produce expected DSN",
			cfg: resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			wantDSN: "host=localhost port=5432 user=postgres password=postgres dbname=barcode sslmode=disable",
		},
		{
			name: "custom values produce correct DSN",
			cfg: resources.Config{
				DBHost:     "db.prod.example.com",
				DBPort:     5433,
				DBUser:     "appuser",
				DBPassword: "s3cr3t!",
				DBName:     "smartcontact",
				DBSSLMode:  "require",
			},
			wantDSN: "host=db.prod.example.com port=5433 user=appuser password=s3cr3t! dbname=smartcontact sslmode=require",
		},
		{
			name: "barcode database name is preserved (migration note)",
			cfg: resources.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "postgres",
				DBName:     "barcode",
				DBSSLMode:  "disable",
			},
			wantDSN: "host=localhost port=5432 user=postgres password=postgres dbname=barcode sslmode=disable",
		},
		{
			name: "sslmode=verify-full is passed through",
			cfg: resources.Config{
				DBHost:     "secure.db",
				DBPort:     5432,
				DBUser:     "u",
				DBPassword: "p",
				DBName:     "d",
				DBSSLMode:  "verify-full",
			},
			wantDSN: "host=secure.db port=5432 user=u password=p dbname=d sslmode=verify-full",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDSN, tt.cfg.DSN())
		})
	}
}

// TestConfig_DSN_ContainsRequiredComponents verifies that the DSN string
// produced from default LoadConfig values contains all mandatory components.
func TestConfig_DSN_ContainsRequiredComponents(t *testing.T) {
	restore := unsetEnv(t, allEnvKeys...)
	defer restore()

	cfg, err := resources.LoadConfig()
	require.NoError(t, err)

	dsn := cfg.DSN()

	components := []struct {
		name  string
		token string
	}{
		{"contains host keyword", "host="},
		{"contains port keyword", "port="},
		{"contains user keyword", "user="},
		{"contains password keyword", "password="},
		{"contains dbname keyword", "dbname="},
		{"contains s