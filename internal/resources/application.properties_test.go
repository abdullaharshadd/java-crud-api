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
// Constants
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, 8082, resources.DefaultServerPort, "DefaultServerPort must mirror server.port=8082")
	assert.Equal(t, "localhost", resources.DefaultDBHost, "DefaultDBHost must be localhost")
	assert.Equal(t, 3306, resources.DefaultDBPort, "DefaultDBPort must be 3306")
	assert.Equal(t, "barcode", resources.DefaultDBName, "DefaultDBName must be barcode")
	assert.Equal(t, "root", resources.DefaultDBUser, "DefaultDBUser must be root")
	assert.Equal(t, "root", resources.DefaultDBPassword, "DefaultDBPassword must be root")
}

func TestDDLAutoMigrationRequired(t *testing.T) {
	// spring.jpa.hibernate.ddl-auto=update has no Go equivalent.
	// The constant must be true to document the migration requirement.
	assert.True(t, resources.DDLAutoMigrationRequired,
		"DDLAutoMigrationRequired must be true to flag schema migration need")
}

// ---------------------------------------------------------------------------
// NewConfig – environment variable handling
// ---------------------------------------------------------------------------

// setenv sets env vars for the duration of a test and restores them on cleanup.
func setenv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		prev, hadPrev := os.LookupEnv(k)
		require.NoError(t, os.Setenv(k, v))
		t.Cleanup(func() {
			if hadPrev {
				_ = os.Setenv(k, prev)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// unsetenv temporarily unsets env vars and restores them on cleanup.
func unsetenv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		prev, hadPrev := os.LookupEnv(k)
		require.NoError(t, os.Unsetenv(k))
		t.Cleanup(func() {
			if hadPrev {
				_ = os.Setenv(k, prev)
			}
		})
	}
}

func TestNewConfig_Defaults(t *testing.T) {
	// Ensure none of the relevant env vars are set.
	unsetenv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD")

	cfg, err := resources.NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, resources.DefaultServerPort, cfg.ServerPort, "default ServerPort")
	assert.Equal(t, resources.DefaultDBHost, cfg.DBHost, "default DBHost")
	assert.Equal(t, resources.DefaultDBPort, cfg.DBPort, "default DBPort")
	assert.Equal(t, resources.DefaultDBName, cfg.DBName, "default DBName")
	assert.Equal(t, resources.DefaultDBUser, cfg.DBUser, "default DBUser")
	assert.Equal(t, resources.DefaultDBPassword, cfg.DBPassword, "default DBPassword")
}

func TestNewConfig_EnvOverrides(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    resources.Config
		wantErr bool
	}{
		{
			name: "all defaults when no env vars set",
			env:  map[string]string{},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
				DBUser:     "root",
				DBPassword: "root",
			},
		},
		{
			name: "server port overridden",
			env:  map[string]string{"SERVER_PORT": "9090"},
			want: resources.Config{
				ServerPort: 9090,
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
				DBUser:     "root",
				DBPassword: "root",
			},
		},
		{
			name: "db host overridden",
			env:  map[string]string{"DB_HOST": "db.example.com"},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "db.example.com",
				DBPort:     3306,
				DBName:     "barcode",
				DBUser:     "root",
				DBPassword: "root",
			},
		},
		{
			name: "db port overridden",
			env:  map[string]string{"DB_PORT": "3307"},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "localhost",
				DBPort:     3307,
				DBName:     "barcode",
				DBUser:     "root",
				DBPassword: "root",
			},
		},
		{
			name: "db name overridden",
			env:  map[string]string{"DB_NAME": "mydb"},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "mydb",
				DBUser:     "root",
				DBPassword: "root",
			},
		},
		{
			name: "db user overridden",
			env:  map[string]string{"DB_USER": "admin"},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
				DBUser:     "admin",
				DBPassword: "root",
			},
		},
		{
			name: "db password overridden",
			env:  map[string]string{"DB_PASSWORD": "s3cr3t"},
			want: resources.Config{
				ServerPort: 8082,
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
				DBUser:     "root",
				DBPassword: "s3cr3t",
			},
		},
		{
			name: "all env vars overridden",
			env: map[string]string{
				"SERVER_PORT": "8080",
				"DB_HOST":     "mysql.prod.local",
				"DB_PORT":     "3308",
				"DB_NAME":     "smartcontact",
				"DB_USER":     "svc_user",
				"DB_PASSWORD": "hunter2",
			},
			want: resources.Config{
				ServerPort: 8080,
				DBHost:     "mysql.prod.local",
				DBPort:     3308,
				DBName:     "smartcontact",
				DBUser:     "svc_user",
				DBPassword: "hunter2",
			},
		},
		{
			name:    "invalid SERVER_PORT returns error",
			env:     map[string]string{"SERVER_PORT": "not-a-number"},
			wantErr: true,
		},
		{
			name:    "invalid DB_PORT returns error",
			env:     map[string]string{"DB_PORT": "abc"},
			wantErr: true,
		},
		{
			name:    "SERVER_PORT with float string returns error",
			env:     map[string]string{"SERVER_PORT": "80.5"},
			wantErr: true,
		},
		{
			name:    "DB_PORT with float string returns error",
			env:     map[string]string{"DB_PORT": "33.06"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Start clean: unset all relevant vars, then apply test-specific ones.
			unsetenv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD")
			if len(tc.env) > 0 {
				setenv(t, tc.env)
			}

			cfg, err := resources.NewConfig()

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tc.want.ServerPort, cfg.ServerPort, "ServerPort")
			assert.Equal(t, tc.want.DBHost, cfg.DBHost, "DBHost")
			assert.Equal(t, tc.want.DBPort, cfg.DBPort, "DBPort")
			assert.Equal(t, tc.want.DBName, cfg.DBName, "DBName")
			assert.Equal(t, tc.want.DBUser, cfg.DBUser, "DBUser")
			assert.Equal(t, tc.want.DBPassword, cfg.DBPassword, "DBPassword")
		})
	}
}

// ---------------------------------------------------------------------------
// NewConfig – error message content
// ---------------------------------------------------------------------------

func TestNewConfig_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		errContains string
	}{
		{
			name:        "SERVER_PORT error wraps key name",
			env:         map[string]string{"SERVER_PORT": "bad"},
			errContains: "SERVER_PORT",
		},
		{
			name:        "DB_PORT error wraps key name",
			env:         map[string]string{"DB_PORT": "bad"},
			errContains: "DB_PORT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			unsetenv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD")
			setenv(t, tc.env)

			_, err := resources.NewConfig()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// ---------------------------------------------------------------------------
// Config.DSN
// ---------------------------------------------------------------------------

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name    string
		cfg     resources.Config
		wantDSN string
	}{
		{
			name: "default values produce correct DSN",
			cfg: resources.Config{
				DBUser:     "root",
				DBPassword: "root",
				DBHost:     "localhost",
				DBPort:     3306,
				DBName:     "barcode",
			},
			wantDSN: "root:root@tcp(localhost:3306)/barcode?parseTime=true",
		},
		{
			name: "custom values produce correct DSN",
			cfg: resources.Config{
				DBUser:     "admin",
				DBPassword: "s3cr3t",
				DBHost:     "db.prod.local",
				DBPort:     3307,
				DBName:     "smartcontact",
			},
			wantDSN: "admin:s3cr3t@tcp(db.prod.local:3307)/smartcontact?parseTime=true",
		},
		{
			name: "DSN includes parseTime=true",
			cfg: resources.Config{
				DBUser:     "u",
				DBPassword: "p",
				DBHost:     "h",
				DBPort:     3306,
				DBName:     "d",
			},
			wantDSN: "u:p@tcp(h:3306)/d?parseTime=true",
		},
		{
			name: "DSN mirrors original JDBC url target: barcode schema on localhost:3306",
			cfg: resources.Config{
				DBUser:     resources.DefaultDBUser,
				DBPassword: resources.DefaultDBPassword,
				DBHost:     resources.DefaultDBHost,
				DBPort:     resources.DefaultDBPort,
				DBName:     resources.DefaultDBName,
			},
			wantDSN: fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?parseTime=true",
				resources.DefaultDBUser,
				resources.DefaultDBPassword,
				resources.DefaultDBHost,
				resources.DefaultDBPort,
				resources.DefaultDBName,
			),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.DSN()
			assert.Equal(t, tc.wantDSN, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Config.ServerAddr
// ---------------------------------------------------------------------------

func TestConfig_ServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		cfg      resources.Config
		wantAddr string
	}{
		{
			name:     "default port 8082",
			cfg:      resources.Config{ServerPort: 8082},
			wantAddr: ":8082",
		},
		{
			name:     "custom port 9090",
			cfg:      resources.Config{ServerPort: 9090},
			wantAddr: ":9090",
		},
		{
			name:     "port 80",
			cfg:      resources.Config{ServerPort: 80},
			wantAddr: ":80",
		},
		{
			name:     "port 443",
			cfg:      resources.Config{ServerPort: 443},
			wantAddr: ":443",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.ServerAddr()
			assert.Equal(t, tc.wantAddr, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Invariant: default config mirrors original application.properties exactly
// ---------------------------------------------------------------------------

func TestNewConfig_DefaultsMirrorOriginalProperties(t *testing.T) {
	unsetenv(t, "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD")

	cfg, err := resources.NewConfig()
	require.NoError(t, err)

	// spring.server.port=8082
	assert.Equal(t, 8082, cfg.ServerPort,
		"server.port invariant: application always serves HTTP on port 8082")

	// spring.datasource.url=jdbc:mysql://localhost:3306/barcode
	assert.Equal(t, "localhost", cfg.DBHost,
		"datasource.url invariant: always targets localhost")
	assert.Equal(t, 3306, cfg.DBPort,
		"datasource.url invariant: always targets port 3306")
	assert.Equal(t, "barcode", cfg.DBName,
		"datasource.url invariant: always targets barcode schema")

	// spring.datasource.username=root
	assert.Equal(t, "root", cfg.DBUser,
		"datasource.username invariant: configured user is always root")

	// spring.datasource.password=root
	assert.Equal(t, "root", cfg.DBPassword,
		"datasource.password invariant: configured password is always root")

	// DSN correctness – combines url + username + password
	assert.Equal(t,
		"root:root@tcp(localhost:3306)/barcode?parseTime=true",
		cfg.DSN(),
		"DSN must mirror JDBC url + credentials from original properties",
	)

	// Server address
	assert.Equal(t, ":8082", cfg.ServerAddr(),
		"ServerAddr must return :8082 for the default configuration")
}

// ---------------------------------------------------------------------------
// Invariant: DDLAutoMigrationRequired signals schema migration is needed
// ---------------------------------------------------------------------------

func TestDDL