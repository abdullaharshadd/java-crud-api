// Package resources contains the application configuration for the
// SmartContact barcode service.
//
// MIGRATION_NOTE: The original source was a Spring Boot
// `application.properties` file. Spring Boot uses externalized configuration
// with auto-configuration: simply declaring `spring.datasource.*` properties
// wires up a DataSource bean, and `spring.jpa.hibernate.ddl-auto=update`
// triggers Hibernate to auto-generate/migrate the schema at startup.
//
// Go has no equivalent auto-wiring or ORM auto-DDL. The idiomatic replacement
// is:
//   - A plain Config struct holding the server port and DB connection details.
//   - A constructor (NewConfig) that reads from environment variables with
//     sensible defaults matching the original property values.
//   - A DSN() method that produces a go-sql-driver/mysql connection string
//     (equivalent to spring.datasource.url + username + password).
//
// Manual review required:
//   - `spring.jpa.hibernate.ddl-auto=update` has NO direct Go equivalent.
//     Schema management must be handled explicitly, e.g. with a migration tool
//     such as golang-migrate or goose. See DDLAutoMigrationRequired below.
//   - `spring.jpa.properties.hibernate.dialect=MySQL8Dialect` is Hibernate
//     specific and is dropped; the MySQL driver handles dialect concerns.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values mirroring the original application.properties.
const (
	// DefaultServerPort mirrors `server.port=8082`.
	DefaultServerPort = 8082

	// DefaultDBHost is derived from `spring.datasource.url` (localhost:3306).
	DefaultDBHost = "localhost"

	// DefaultDBPort is derived from `spring.datasource.url` (localhost:3306).
	DefaultDBPort = 3306

	// DefaultDBName mirrors the schema `barcode` from spring.datasource.url.
	DefaultDBName = "barcode"

	// DefaultDBUser mirrors `spring.datasource.username=root`.
	DefaultDBUser = "root"

	// DefaultDBPassword mirrors `spring.datasource.password=root`.
	//
	// MIGRATION_NOTE: Hardcoding credentials is a security anti-pattern.
	// This default preserves the original file's behavior, but production
	// deployments MUST override it via the DB_PASSWORD environment variable
	// or a secrets manager.
	DefaultDBPassword = "root"
)

// DDLAutoMigrationRequired documents that the source configured Hibernate to
// auto-update the database schema (spring.jpa.hibernate.ddl-auto=update).
//
// MIGRATION_NOTE: There is no Go equivalent of Hibernate DDL auto-generation.
// Schema creation/migration must be performed explicitly (e.g. golang-migrate,
// goose, or hand-written SQL) as part of application startup or deployment.
const DDLAutoMigrationRequired = true

// Config holds the runtime configuration for the SmartContact service.
//
// MIGRATION_NOTE: Replaces Spring Boot's externalized `application.properties`
// configuration. Values are read from environment variables so the same
// binary can be configured per-environment without recompilation.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on.
	ServerPort int

	// DBHost is the MySQL server hostname.
	DBHost string

	// DBPort is the MySQL server port.
	DBPort int

	// DBName is the MySQL database (schema) name.
	DBName string

	// DBUser is the MySQL username.
	DBUser string

	// DBPassword is the MySQL password.
	DBPassword string
}

// NewConfig builds a Config from environment variables, falling back to the
// defaults that mirror the original application.properties file.
//
// It returns an error if a provided numeric environment variable
// (SERVER_PORT or DB_PORT) cannot be parsed as an integer.
func NewConfig() (*Config, error) {
	serverPort, err := intFromEnv("SERVER_PORT", DefaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("parse SERVER_PORT: %w", err)
	}

	dbPort, err := intFromEnv("DB_PORT", DefaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("parse DB_PORT: %w", err)
	}

	return &Config{
		ServerPort: serverPort,
		DBHost:     stringFromEnv("DB_HOST", DefaultDBHost),
		DBPort:     dbPort,
		DBName:     stringFromEnv("DB_NAME", DefaultDBName),
		DBUser:     stringFromEnv("DB_USER", DefaultDBUser),
		DBPassword: stringFromEnv("DB_PASSWORD", DefaultDBPassword),
	}, nil
}

// DSN returns the MySQL data source name for use with the
// github.com/go-sql-driver/mysql driver (via database/sql or sqlx).
//
// MIGRATION_NOTE: This replaces the JDBC URL
// `jdbc:mysql://localhost:3306/barcode` together with the username/password
// properties. parseTime=true is added so DATE/DATETIME columns scan into
// time.Time, which Hibernate handled implicitly.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// ServerAddr returns the listen address for the HTTP server, e.g. ":8082".
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// stringFromEnv returns the value of the named environment variable, or the
// provided fallback when the variable is unset or empty.
func stringFromEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// intFromEnv returns the integer value of the named environment variable, or
// the provided fallback when the variable is unset or empty. It returns an
// error if the variable is set but cannot be parsed as an integer.
func intFromEnv(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q for %s: %w", v, key, err)
	}
	return n, nil
}
