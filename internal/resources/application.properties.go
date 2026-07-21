// Package resources holds application configuration for the SmartContact
// service. It is the Go equivalent of the Spring Boot
// src/main/resources/application.properties file.
//
// MIGRATION_NOTE: The source application.properties was a Spring Boot
// externalized-configuration file. Spring Boot's auto-configuration read
// these keys at startup to build a DataSource and an
// EntityManagerFactory, and Hibernate managed the schema via
// ddl-auto=update.
//
// Go has no equivalent auto-configuration mechanism, so this file does
// NOT reproduce Spring's magic. Instead it provides:
//
//   - A plain Config struct holding the same settings as typed Go fields.
//   - Sensible defaults matching the original property values.
//   - Environment-variable overrides, which is the idiomatic Go
//     (12-factor) way to externalize configuration. Hard-coding
//     credentials like "root/root" in source is a security smell; the
//     defaults are kept only to preserve behaviour, but production
//     deployments MUST override them via the environment.
//   - A DSN() helper that produces a go-sql-driver/mysql connection
//     string, replacing the JDBC URL and driver-class-name.
//
// MIGRATION_NOTE: The following Spring/Hibernate concepts have NO direct
// Go equivalent and require manual review:
//   - spring.datasource.driver-class-name: In Go the driver is selected
//     by importing github.com/go-sql-driver/mysql and passing "mysql"
//     to sql.Open; there is no class name.
//   - spring.jpa.hibernate.ddl-auto=update: Go has no ORM that
//     auto-migrates by default. Schema management must be done
//     explicitly, e.g. with golang-migrate or goose. This field is
//     preserved for documentation only and is not acted upon here.
//   - spring.jpa.properties.hibernate.dialect: SQL dialect handling is
//     the responsibility of the chosen driver; there is no dialect knob.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values mirror the original application.properties.
const (
	defaultServerPort = 8082
	defaultDBHost     = "localhost"
	defaultDBPort     = 3306
	defaultDBName     = "barcode"
	defaultDBUser     = "root"
	defaultDBPassword = "root"

	// defaultDDLAuto preserves the original Hibernate ddl-auto=update
	// setting for documentation purposes. It is not enforced by Go.
	defaultDDLAuto = "update"
)

// Config holds the runtime configuration for the SmartContact service.
// It is the typed Go replacement for the Spring Boot
// application.properties file.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on.
	// Replaces server.port.
	ServerPort int

	// DBHost is the MySQL server host. Derived from
	// spring.datasource.url.
	DBHost string

	// DBPort is the MySQL server port. Derived from
	// spring.datasource.url.
	DBPort int

	// DBName is the MySQL database (schema) name. Derived from
	// spring.datasource.url.
	DBName string

	// DBUser is the MySQL username. Replaces
	// spring.datasource.username.
	DBUser string

	// DBPassword is the MySQL password. Replaces
	// spring.datasource.password.
	//
	// MIGRATION_NOTE: Storing a password default in source is unsafe;
	// override via the DB_PASSWORD environment variable in production.
	DBPassword string

	// DDLAuto records the original Hibernate ddl-auto value. Go performs
	// no automatic schema migration; use a migration tool instead.
	DDLAuto string
}

// NewConfig returns a Config populated from environment variables,
// falling back to the defaults that mirror the original
// application.properties. It never returns an error for missing
// variables (defaults are used); it returns an error only when an
// override cannot be parsed into the expected type.
func NewConfig() (*Config, error) {
	serverPort, err := envInt("SERVER_PORT", defaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("parsing SERVER_PORT: %w", err)
	}

	dbPort, err := envInt("DB_PORT", defaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("parsing DB_PORT: %w", err)
	}

	return &Config{
		ServerPort: serverPort,
		DBHost:     envString("DB_HOST", defaultDBHost),
		DBPort:     dbPort,
		DBName:     envString("DB_NAME", defaultDBName),
		DBUser:     envString("DB_USER", defaultDBUser),
		DBPassword: envString("DB_PASSWORD", defaultDBPassword),
		DDLAuto:    envString("DB_DDL_AUTO", defaultDDLAuto),
	}, nil
}

// Addr returns the HTTP server listen address in ":port" form, suitable
// for passing to http.Server.Addr or net.Listen.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// DSN returns a github.com/go-sql-driver/mysql data source name that is
// equivalent to the original JDBC URL
// (jdbc:mysql://host:port/db). Pass the result to sql.Open("mysql", dsn).
//
// parseTime=true is enabled so that MySQL DATE/DATETIME columns scan
// into Go time.Time values, matching the behaviour applications usually
// expect from Hibernate-mapped temporal fields.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// envString returns the value of the named environment variable, or the
// provided fallback if the variable is unset or empty.
func envString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// envInt returns the integer value of the named environment variable, or
// the provided fallback if the variable is unset or empty. It returns an
// error if the variable is set but cannot be parsed as an integer.
func envInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", v, err)
	}
	return n, nil
}
