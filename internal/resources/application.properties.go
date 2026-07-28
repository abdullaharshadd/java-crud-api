// Package resources documents and provides the runtime configuration for
// the SmartContact application.
//
// MIGRATION_NOTE: The source file (application.properties) was a Spring Boot
// externalized-configuration file. Spring Boot's auto-configuration binds these
// properties to the embedded server, the DataSource, and JPA/Hibernate at
// startup. Go has no equivalent convention-over-configuration magic, so this
// file translates the *intent* of each property into an explicit, typed
// configuration struct plus a constructor that reads from environment
// variables (the idiomatic Go analogue of externalized config).
//
// MIGRATION_NOTE — DATABASE DIALECT: The source targeted MySQL
// (jdbc:mysql, com.mysql.cj.jdbc.Driver, MySQL8Dialect). The migration target
// is PostgreSQL. The DSN and driver below have been deliberately re-shaped for
// PostgreSQL rather than mirroring the MySQL source. Do NOT reintroduce the
// MySQL connection string.
//
// MIGRATION_NOTE — SCHEMA MANAGEMENT: spring.jpa.hibernate.ddl-auto=update told
// Hibernate to auto-create/alter tables from the entity model on startup. Go
// has no ORM performing this; schema is managed explicitly via SQL migrations
// (see migrations/0001_create_users.sql). This behaviour must be reproduced by
// running migrations at deploy time — it is intentionally NOT done implicitly
// at application startup.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values mirroring the original application.properties.
//
// MIGRATION_NOTE: The original username/password were the literal "root"/"root"
// development credentials. These MUST NOT be shipped as production defaults;
// they are retained only as local-development fallbacks and should always be
// overridden via environment variables in any real deployment.
const (
	defaultServerPort = 8082
	defaultDBHost     = "localhost"
	// MIGRATION_NOTE: MySQL's default port was 3306; PostgreSQL's is 5432.
	defaultDBPort     = 5432
	defaultDBName     = "barcode"
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBSSLMode  = "disable"
)

// Config holds the runtime configuration for the SmartContact application.
// It is the idiomatic Go replacement for Spring Boot's application.properties.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on.
	// Source: server.port.
	ServerPort int

	// DBHost is the database server hostname.
	DBHost string

	// DBPort is the database server port.
	DBPort int

	// DBName is the target database/schema name.
	// Source: the database segment of spring.datasource.url.
	DBName string

	// DBUser is the database username.
	// Source: spring.datasource.username.
	DBUser string

	// DBPassword is the database password.
	// Source: spring.datasource.password.
	DBPassword string

	// DBSSLMode is the PostgreSQL sslmode connection parameter.
	// MIGRATION_NOTE: No MySQL equivalent existed in the source; "disable"
	// is a safe local-development default. Set to "require" (or stricter)
	// in production.
	DBSSLMode string
}

// LoadConfig builds a Config from environment variables, falling back to the
// development defaults derived from the original application.properties.
//
// Supported environment variables:
//
//	SERVER_PORT, DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
//
// It returns an error if a provided numeric variable cannot be parsed.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort: defaultServerPort,
		DBHost:     defaultDBHost,
		DBPort:     defaultDBPort,
		DBName:     defaultDBName,
		DBUser:     defaultDBUser,
		DBPassword: defaultDBPassword,
		DBSSLMode:  defaultDBSSLMode,
	}

	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVER_PORT %q: %w", v, err)
		}
		cfg.ServerPort = port
	}

	if v, ok := os.LookupEnv("DB_HOST"); ok {
		cfg.DBHost = v
	}

	if v, ok := os.LookupEnv("DB_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_PORT %q: %w", v, err)
		}
		cfg.DBPort = port
	}

	if v, ok := os.LookupEnv("DB_NAME"); ok {
		cfg.DBName = v
	}

	if v, ok := os.LookupEnv("DB_USER"); ok {
		cfg.DBUser = v
	}

	if v, ok := os.LookupEnv("DB_PASSWORD"); ok {
		cfg.DBPassword = v
	}

	if v, ok := os.LookupEnv("DB_SSLMODE"); ok {
		cfg.DBSSLMode = v
	}

	return cfg, nil
}

// DSN returns a PostgreSQL connection string (key/value form) suitable for
// sql.Open with the lib/pq or pgx stdlib driver.
//
// MIGRATION_NOTE: This replaces the MySQL JDBC URL
// (jdbc:mysql://localhost:3306/barcode). The target dialect is PostgreSQL, so
// the connection string is built in PostgreSQL's native form rather than a
// translated JDBC URL.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// ServerAddr returns the ":port" address string for use with http.Server.Addr.
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}
