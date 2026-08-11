// Package resources holds application configuration migrated from Spring
// Boot's src/main/resources/application.properties.
//
// MIGRATION_NOTE: The Java source was a Spring Boot application.properties
// file. Spring Boot loads these key/value pairs at startup via its
// externalized-configuration mechanism and auto-wires the embedded server
// port, the JDBC datasource, and Hibernate's ddl-auto schema management.
//
// Go has no equivalent auto-configuration container. The idiomatic approach
// is to read configuration from environment variables at startup, falling
// back to sensible defaults. The individual settings map to Go as follows:
//
//   - server.port                        -> Config.ServerAddr (SERVER_PORT env)
//   - spring.datasource.url/username/... -> Config.DatabaseURL (DATABASE_URL env)
//   - spring.jpa.hibernate.ddl-auto      -> handled explicitly by
//                                           model.EnsureUserSchema at startup
//                                           (see smartcontactapplication.go);
//                                           there is no runtime ORM DDL.
//   - spring.jpa.properties.hibernate.dialect -> no equivalent; the SQL in
//                                           the repository layer is written
//                                           directly for the target dialect.
//
// CRITICAL: the source targeted MySQL, but the target database for this
// migration is PostgreSQL. The default DSN below is a PostgreSQL DSN and the
// database name ("barcode") is preserved from the source. Credentials must be
// supplied via the DATABASE_URL environment variable in any real deployment;
// the hard-coded root/root defaults from the source are development-only.
package resources

import (
	"fmt"
	"os"
)

const (
	// DefaultServerPort mirrors server.port=8082 from the source properties.
	DefaultServerPort = "8082"

	// DefaultDatabaseURL is the fallback PostgreSQL DSN used when the
	// DATABASE_URL environment variable is not set.
	//
	// MIGRATION_NOTE: the source used a MySQL JDBC URL
	// (jdbc:mysql://localhost:3306/barcode with root/root). Per the target
	// dialect this is translated to a PostgreSQL connection string. The
	// database name "barcode" is preserved.
	DefaultDatabaseURL = "postgres://root:root@localhost:5432/barcode?sslmode=disable"
)

// Config holds the runtime configuration for the SmartContact application.
// It replaces the Spring Boot application.properties externalized config.
type Config struct {
	// ServerAddr is the address the HTTP server listens on, e.g. ":8082".
	ServerAddr string

	// DatabaseURL is the PostgreSQL connection string (DSN) used to open the
	// database handle.
	DatabaseURL string
}

// LoadConfig builds a Config from the environment, applying the defaults
// derived from the source application.properties when a variable is unset.
//
// Recognised environment variables:
//
//	SERVER_PORT   - the port to listen on (default "8082").
//	DATABASE_URL  - the full PostgreSQL DSN (default DefaultDatabaseURL).
//
// It never returns an error today, but returns one to keep the signature
// stable for future validation (e.g. required-variable enforcement).
func LoadConfig() (*Config, error) {
	port := getenvDefault("SERVER_PORT", DefaultServerPort)
	dbURL := getenvDefault("DATABASE_URL", DefaultDatabaseURL)

	return &Config{
		ServerAddr:  fmt.Sprintf(":%s", port),
		DatabaseURL: dbURL,
	}, nil
}

// getenvDefault returns the value of the environment variable named by key,
// or fallback if the variable is unset or empty.
func getenvDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
