// Package resources holds the migrated application configuration.
//
// MIGRATION_NOTE: The Java source was a Spring Boot application.properties file
// that Spring's externalized-configuration machinery read implicitly at boot to
// wire the embedded server port, the MySQL DataSource, and Hibernate's
// DDL/dialect behavior. Go has no equivalent auto-binding of a properties file
// to framework internals, so that implicit configuration becomes an explicit,
// typed config loader (config.Load) that reads from environment variables with
// sane defaults derived from the original property values.
//
// Key mappings from the source properties:
//
//	server.port=8082                         -> Config.ServerPort  (SERVER_PORT)
//	spring.datasource.url                    -> Config.DatabaseURL (DATABASE_URL)
//	spring.datasource.username=root          -> folded into DATABASE_URL
//	spring.datasource.password=root          -> folded into DATABASE_URL
//	spring.datasource.driver-class-name      -> dropped; the driver is chosen at
//	                                            compile time by the Postgres
//	                                            database/sql import, not config.
//	spring.jpa.hibernate.ddl-auto=update     -> handled by db.Connect's
//	                                            ensureSchema (CREATE TABLE IF NOT
//	                                            EXISTS), not a property.
//	spring.jpa.properties.hibernate.dialect  -> dropped; the target dialect is
//	                                            PostgreSQL, wired via the driver.
//
// IMPORTANT — dialect change: the source targeted MySQL 8 (jdbc:mysql,
// MySQL8Dialect). Per the migration's target-database directive this project
// standardizes on PostgreSQL, so the default DSN below is a PostgreSQL DSN.
// This is an intentional, project-wide decision — not a mechanical port of the
// MySQL URL.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values, derived from the original
// application.properties. The database default is expressed as a PostgreSQL
// DSN because the target database dialect for this migration is PostgreSQL.
const (
	// DefaultServerPort mirrors the source's server.port=8082.
	DefaultServerPort = 8082

	// DefaultDatabaseURL is the PostgreSQL equivalent of the source's
	// jdbc:mysql://localhost:3306/barcode with username/password root/root.
	DefaultDatabaseURL = "postgres://root:root@localhost:5432/barcode?sslmode=disable"
)

// Environment variable names used to override the defaults above.
const (
	EnvServerPort  = "SERVER_PORT"
	EnvDatabaseURL = "DATABASE_URL"
)

// Config holds the runtime configuration for the smartcontact application.
//
// It is the typed, explicit replacement for the implicit Spring Boot
// externalized configuration that the original application.properties provided.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on.
	ServerPort int

	// DatabaseURL is the PostgreSQL connection string (DSN) used to open the
	// database/sql pool.
	DatabaseURL string
}

// Addr returns the HTTP listen address (e.g. ":8082") suitable for
// http.Server.Addr or net.Listen.
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// Load builds a Config from environment variables, falling back to the defaults
// derived from the original application.properties when a variable is unset or
// empty.
//
// It returns an error if an override is present but invalid (for example a
// non-numeric SERVER_PORT), so misconfiguration fails fast at startup rather
// than surfacing as an obscure runtime failure.
func Load() (Config, error) {
	cfg := Config{
		ServerPort:  DefaultServerPort,
		DatabaseURL: DefaultDatabaseURL,
	}

	if raw := os.Getenv(EnvServerPort); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("resources: invalid %s %q: %w", EnvServerPort, raw, err)
		}
		if port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("resources: %s %d out of range (1-65535)", EnvServerPort, port)
		}
		cfg.ServerPort = port
	}

	if raw := os.Getenv(EnvDatabaseURL); raw != "" {
		cfg.DatabaseURL = raw
	}

	return cfg, nil
}
