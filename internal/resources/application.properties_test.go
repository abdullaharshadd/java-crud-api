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
// Constants
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	t.Run("DefaultServerPort is 8082", func(t *testing.T) {
		assert.Equal(t, "8082", resources.DefaultServerPort,
			"server.port must default to 8082 (mirrors Spring Boot application.properties)")
	})

	t.Run("DefaultDatabaseURL targets barcode database", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "barcode",
			"default DSN must reference the 'barcode' database")
	})

	t.Run("DefaultDatabaseURL contains root username", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "root",
			"default DSN must embed username 'root' (migrated from spring.datasource.username)")
	})

	t.Run("DefaultDatabaseURL contains root password", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "root:root",
			"default DSN must embed password 'root' (migrated from spring.datasource.password)")
	})

	t.Run("DefaultDatabaseURL is PostgreSQL DSN", func(t *testing.T) {
		assert.True(t,
			strings.HasPrefix(resources.DefaultDatabaseURL, "postgres://"),
			"target dialect is PostgreSQL; DSN must start with postgres://")
	})

	t.Run("DefaultDatabaseURL contains localhost:5432", func(t *testing.T) {
		assert.Contains(t, resources.DefaultDatabaseURL, "localhost:5432",
			"default DSN must point to localhost PostgreSQL port")
	})
}

// ---------------------------------------------------------------------------
// LoadConfig – table-driven
// ---------------------------------------------------------------------------

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		envServerPort   string // empty string = unset
		envDatabaseURL  string // empty string = unset
		wantServerAddr  string
		wantDatabaseURL string
		wantErr         bool
	}{
		{
			name:            "all defaults when no env vars set",
			envServerPort:   "",
			envDatabaseURL:  "",
			wantServerAddr:  fmt.Sprintf(":%s", resources.DefaultServerPort),
			wantDatabaseURL: resources.DefaultDatabaseURL,
			wantErr:         false,
		},
		{
			name:            "SERVER_PORT overrides default port",
			envServerPort:   "9090",
			envDatabaseURL:  "",
			wantServerAddr:  ":9090",
			wantDatabaseURL: resources.DefaultDatabaseURL,
			wantErr:         false,
		},
		{
			name:            "DATABASE_URL overrides default DSN",
			envServerPort:   "",
			envDatabaseURL:  "postgres://user:pass@db.example.com:5432/mydb?sslmode=require",
			wantServerAddr:  fmt.Sprintf(":%s", resources.DefaultServerPort),
			wantDatabaseURL: "postgres://user:pass@db.example.com:5432/mydb?sslmode=require",
			wantErr:         false,
		},
		{
			name:            "both env vars override defaults",
			envServerPort:   "8000",
			envDatabaseURL:  "postgres://admin:secret@remote:5432/prod",
			wantServerAddr:  ":8000",
			wantDatabaseURL: "postgres://admin:secret@remote:5432/prod",
			wantErr:         false,
		},
		{
			name:            "SERVER_PORT=8082 matches spring boot default",
			envServerPort:   "8082",
			envDatabaseURL:  "",
			wantServerAddr:  ":8082",
			wantDatabaseURL: resources.DefaultDatabaseURL,
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Isolate env changes to this sub-test.
			setOrUnset(t, "SERVER_PORT", tc.envServerPort)
			setOrUnset(t, "DATABASE_URL", tc.envDatabaseURL)

			cfg, err := resources.LoadConfig()

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.wantServerAddr, cfg.ServerAddr,
				"ServerAddr should reflect SERVER_PORT env or default")
			assert.Equal(t, tc.wantDatabaseURL, cfg.DatabaseURL,
				"DatabaseURL should reflect DATABASE_URL env or default")
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: server.port – application binds to port 8082
// ---------------------------------------------------------------------------

func TestServerPort(t *testing.T) {
	tests := []struct {
		name           string
		serverPortEnv  string
		wantServerAddr string
	}{
		{
			name:           "default port is 8082 (spring server.port invariant)",
			serverPortEnv:  "",
			wantServerAddr: ":8082",
		},
		{
			name:           "explicit 8082 env still produces :8082",
			serverPortEnv:  "8082",
			wantServerAddr: ":8082",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "SERVER_PORT", tc.serverPortEnv)
			setOrUnset(t, "DATABASE_URL", "")

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantServerAddr, cfg.ServerAddr)
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: spring.datasource – username, password, database name
// ---------------------------------------------------------------------------

func TestDatasourceCredentialsAndDatabase(t *testing.T) {
	tests := []struct {
		name            string
		databaseURLEnv  string
		wantUser        bool   // DSN contains "root" as user
		wantPassword    bool   // DSN contains "root" as password
		wantDatabase    string // expected substring in DSN
		wantDSNContains string
	}{
		{
			name:            "default DSN embeds root:root credentials (spring username/password)",
			databaseURLEnv:  "",
			wantUser:        true,
			wantPassword:    true,
			wantDatabase:    "barcode",
			wantDSNContains: "root:root",
		},
		{
			name:            "default DSN targets barcode database (spring.datasource.url invariant)",
			databaseURLEnv:  "",
			wantDatabase:    "barcode",
			wantDSNContains: "barcode",
		},
		{
			name:           "custom DATABASE_URL overrides all defaults",
			databaseURLEnv: "postgres://admin:s3cr3t@prod:5432/barcode?sslmode=require",
			wantDatabase:   "barcode",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "DATABASE_URL", tc.databaseURLEnv)
			setOrUnset(t, "SERVER_PORT", "")

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tc.wantDatabase != "" {
				assert.Contains(t, cfg.DatabaseURL, tc.wantDatabase,
					"DatabaseURL must reference the 'barcode' database")
			}
			if tc.wantDSNContains != "" {
				assert.Contains(t, cfg.DatabaseURL, tc.wantDSNContains,
					"DatabaseURL must contain expected substring")
			}
			if tc.wantUser && tc.databaseURLEnv == "" {
				assert.Contains(t, cfg.DatabaseURL, "root",
					"default DSN must embed 'root' as username")
			}
			if tc.wantPassword && tc.databaseURLEnv == "" {
				assert.Contains(t, cfg.DatabaseURL, "root:root",
					"default DSN must embed 'root' as password")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: spring.datasource.url – localhost host reference
// ---------------------------------------------------------------------------

func TestDatasourceURL(t *testing.T) {
	tests := []struct {
		name            string
		databaseURLEnv  string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:           "default DSN targets localhost (migrated from localhost:3306)",
			databaseURLEnv: "",
			wantContains:   []string{"localhost", "barcode"},
		},
		{
			name:           "default DSN is a PostgreSQL DSN not MySQL JDBC URL",
			databaseURLEnv: "",
			wantContains:   []string{"postgres://"},
			wantNotContains: []string{
				"jdbc:mysql",
				"com.mysql",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "DATABASE_URL", tc.databaseURLEnv)
			setOrUnset(t, "SERVER_PORT", "")

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)

			for _, want := range tc.wantContains {
				assert.Contains(t, cfg.DatabaseURL, want)
			}
			for _, notWant := range tc.wantNotContains {
				assert.NotContains(t, cfg.DatabaseURL, notWant)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: spring.jpa.hibernate.ddl-auto – schema management migrated to
// EnsureUserSchema; LoadConfig itself should not alter schema.
// The test verifies the configuration does not carry a conflicting DDL
// setting and that Config is structurally correct for downstream use.
// ---------------------------------------------------------------------------

func TestDDLAutoMigrationNote(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "LoadConfig returns a valid Config (DDL is handled externally)"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "SERVER_PORT", "")
			setOrUnset(t, "DATABASE_URL", "")

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg,
				"LoadConfig must return a non-nil Config; DDL is managed by model.EnsureUserSchema")
			assert.NotEmpty(t, cfg.ServerAddr)
			assert.NotEmpty(t, cfg.DatabaseURL)
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: spring.jpa.properties.hibernate.dialect – no MySQL dialect in target.
// The migrated DSN must not contain MySQL-specific markers.
// ---------------------------------------------------------------------------

func TestHibernateDialectMigration(t *testing.T) {
	tests := []struct {
		name            string
		databaseURLEnv  string
		wantNotContains []string
	}{
		{
			name:           "default DSN has no MySQL dialect artifacts",
			databaseURLEnv: "",
			wantNotContains: []string{
				"mysql",
				"MySQL8Dialect",
				"jdbc",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "DATABASE_URL", tc.databaseURLEnv)
			setOrUnset(t, "SERVER_PORT", "")

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)

			for _, notWant := range tc.wantNotContains {
				assert.NotContains(t,
					strings.ToLower(cfg.DatabaseURL),
					strings.ToLower(notWant),
					"migrated DSN must not reference MySQL-specific tokens")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: empty / whitespace env var falls back to default
// ---------------------------------------------------------------------------

func TestGetenvDefaultFallback(t *testing.T) {
	tests := []struct {
		name            string
		serverPortEnv   string
		databaseURLEnv  string
		wantServerAddr  string
		wantDatabaseURL string
	}{
		{
			name:            "empty SERVER_PORT falls back to 8082",
			serverPortEnv:   "",
			databaseURLEnv:  "",
			wantServerAddr:  ":8082",
			wantDatabaseURL: resources.DefaultDatabaseURL,
		},
		{
			name:            "empty DATABASE_URL falls back to DefaultDatabaseURL",
			serverPortEnv:   "",
			databaseURLEnv:  "",
			wantServerAddr:  ":8082",
			wantDatabaseURL: resources.DefaultDatabaseURL,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Explicitly set to empty string to simulate unset.
			setOrUnset(t, "SERVER_PORT", tc.serverPortEnv)
			setOrUnset(t, "DATABASE_URL", tc.databaseURLEnv)

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.wantServerAddr, cfg.ServerAddr)
			assert.Equal(t, tc.wantDatabaseURL, cfg.DatabaseURL)
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: Config struct fields are populated correctly
// ---------------------------------------------------------------------------

func TestConfigStructFields(t *testing.T) {
	tests := []struct {
		name           string
		serverPortEnv  string
		databaseURLEnv string
		wantServerAddr string
		wantDB         string
	}{
		{
			name:           "ServerAddr uses colon-prefix format",
			serverPortEnv:  "3000",
			databaseURLEnv: "",
			wantServerAddr: ":3000",
			wantDB:         resources.DefaultDatabaseURL,
		},
		{
			name:           "DatabaseURL is stored verbatim",
			serverPortEnv:  "",
			databaseURLEnv: "postgres://u:p@h:5432/d",
			wantServerAddr: ":8082",
			wantDB:         "postgres://u:p@h:5432/d",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setOrUnset(t, "SERVER_PORT", tc.serverPortEnv)
			setOrUnset(t, "DATABASE_URL", tc.databaseURLEnv)

			cfg, err := resources.LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.wantServerAddr, cfg.ServerAddr)
			assert.Equal(t, tc.wantDB, cfg.DatabaseURL)
		})
	}
}

// ---------------------------------------------------------------------------
// Invariant: LoadConfig never returns nil Config on success
// ---------------------------------------------------------------------------

func TestLoadConfigNeverReturnsNilConfigOnSuccess(t *testing.T) {
	setOrUnset(t, "SERVER_PORT", "")
	setOrUnset(t, "DATABASE_URL", "