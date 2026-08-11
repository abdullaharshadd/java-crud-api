package resources

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
)

// Default configuration values. These mirror the development defaults that
// were hard-coded in the source application.properties. Production deployments
// should override every value via environment variables.
const (
	defaultServerPort = 8082
	defaultDBHost     = "localhost"
	defaultDBPort     = 5432 // PostgreSQL default; source used MySQL's 3306
	defaultDBName     = "barcode"
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBSSLMode  = "disable"
)

// createUsersTableSQL is the DDL for the users table, derived from the User
// model in internal/smartcontact/model/user.go. Column names match the `db`
// struct tags exactly.
const createUsersTableSQL = `
CREATE TABLE IF NOT EXISTS users (
    user_id       SERIAL PRIMARY KEY,
    user_name     TEXT NOT NULL,
    user_email    TEXT NOT NULL UNIQUE,
    user_password TEXT NOT NULL,
    user_role     TEXT NOT NULL DEFAULT '',
    user_about    TEXT NOT NULL DEFAULT ''
)`

// Config holds all runtime configuration for the Smart Contact service. It is
// the idiomatic Go replacement for Spring Boot's externalized properties.
type Config struct {
	// ServerPort is the TCP port the embedded HTTP server listens on.
	ServerPort int

	// DBHost is the PostgreSQL server hostname.
	DBHost string
	// DBPort is the PostgreSQL server port.
	DBPort int
	// DBName is the target database/schema name.
	DBName string
	// DBUser is the database user.
	DBUser string
	// DBPassword is the database password.
	DBPassword string
	// DBSSLMode is the lib/pq sslmode parameter (disable, require, etc.).
	DBSSLMode string
}

// LoadConfig builds a Config from environment variables, falling back to the
// development defaults for any variable that is unset. It returns an error if
// a variable that is expected to be numeric cannot be parsed.
//
// Recognized environment variables:
//
//	SERVER_PORT, DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
//
// Also recognizes DATABASE_URL as a complete DSN override (takes precedence
// over individual DB_* variables when set).
func LoadConfig() (*Config, error) {
	serverPort, err := intFromEnv("SERVER_PORT", defaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("loading SERVER_PORT: %w", err)
	}

	dbPort, err := intFromEnv("DB_PORT", defaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("loading DB_PORT: %w", err)
	}

	cfg := &Config{
		ServerPort: serverPort,
		DBHost:     stringFromEnv("DB_HOST", defaultDBHost),
		DBPort:     dbPort,
		DBName:     stringFromEnv("DB_NAME", stringFromEnv("DB_DATABASE", stringFromEnv("POSTGRES_DB", defaultDBName))),
		DBUser:     stringFromEnv("DB_USER", stringFromEnv("DB_USERNAME", stringFromEnv("POSTGRES_USER", defaultDBUser))),
		DBPassword: stringFromEnv("DB_PASSWORD", stringFromEnv("POSTGRES_PASSWORD", defaultDBPassword)),
		DBSSLMode:  stringFromEnv("DB_SSLMODE", defaultDBSSLMode),
	}
	return cfg, nil
}

// DSN returns a PostgreSQL data source name suitable for sql.Open("postgres", ...)
// with the lib/pq driver. This replaces Spring's spring.datasource.url.
//
// If the DATABASE_URL environment variable is set, it is returned as-is and
// takes precedence over the individual DB_* fields.
func (c *Config) DSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Addr returns the host:port address the HTTP server should bind to, suitable
// for http.Server.Addr. This replaces Spring's server.port.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// EnsureSchema creates the users table (and any other required tables) in the
// database if they do not already exist. This replaces the JPA
// hibernate.ddl-auto=update behaviour from the source application.properties.
// It must be called once after the *sql.DB is opened and before any queries
// are executed.
func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		return fmt.Errorf("EnsureSchema: creating users table: %w", err)
	}
	return nil
}

// stringFromEnv returns the value of the named environment variable, or the
// provided default if the variable is unset or empty.
func stringFromEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// intFromEnv returns the integer value of the named environment variable, or
// the provided default if the variable is unset or empty. It returns an error
// if the variable is set but not a valid integer.
func intFromEnv(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q: %w", key, v, err)
	}
	return n, nil
}