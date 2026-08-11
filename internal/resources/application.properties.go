// Package resources holds runtime configuration for the Smart Contact
// service. It is the Go equivalent of the source project's Spring Boot
// application.properties file.
//
// MIGRATION_NOTE: The Java source was a Spring Boot application.properties
// file. Spring Boot's externalized configuration and auto-configuration have
// no direct Go equivalent — there is no framework that reflectively wires a
// DataSource or JPA EntityManager from key/value properties. In idiomatic Go
// we load configuration explicitly from environment variables (with sane
// defaults) into a typed Config struct, and the caller uses that struct to
// build an *sql.DB and an *http.Server.
//
// Key translation decisions:
//
//   - server.port                  -> Config.ServerPort (env SERVER_PORT)
//   - spring.datasource.url        -> decomposed into host/port/name and
//                                     rebuilt as a PostgreSQL DSN. The source
//                                     used MySQL (jdbc:mysql://.../barcode),
//                                     but the TARGET database is PostgreSQL,
//                                     so we deliberately emit a lib/pq-style
//                                     DSN and do NOT mirror the MySQL dialect.
//   - spring.datasource.username   -> Config.DBUser (env DB_USER)
//   - spring.datasource.password   -> Config.DBPassword (env DB_PASSWORD)
//   - spring.datasource.driver...  -> dropped; the Go driver is chosen at
//                                     import time (e.g. _ "github.com/lib/pq").
//   - spring.jpa.hibernate.ddl-auto=update -> no equivalent. Schema migrations
//                                     should be handled explicitly (e.g. with
//                                     golang-migrate) rather than auto-applied
//                                     at startup. Flagged for manual review.
//   - spring.jpa...dialect=MySQL8  -> intentionally NOT carried over; PostgreSQL
//                                     is the target dialect.
//
// The hard-coded credentials in the source (root/root) are treated as
// development-only defaults here; production deployments MUST override them
// via environment variables.
package resources

import (
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
		DBName:     stringFromEnv("DB_NAME", defaultDBName),
		DBUser:     stringFromEnv("DB_USER", defaultDBUser),
		DBPassword: stringFromEnv("DB_PASSWORD", defaultDBPassword),
		DBSSLMode:  stringFromEnv("DB_SSLMODE", defaultDBSSLMode),
	}
	return cfg, nil
}

// DSN returns a PostgreSQL data source name suitable for sql.Open("postgres", ...)
// with the lib/pq driver. This replaces Spring's spring.datasource.url.
func (c *Config) DSN() string {
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
