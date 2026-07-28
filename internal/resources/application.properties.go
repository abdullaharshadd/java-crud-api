package resources

// Package resources provides application configuration loading for the
// SmartContact service. It replaces the Java Spring Boot externalized
// configuration file src/main/resources/application.properties.
//
// MIGRATION_NOTE: The original application.properties relied on Spring Boot's
// auto-configuration to build a DataSource and JPA/Hibernate EntityManager from
// a handful of properties (server.port, spring.datasource.*, spring.jpa.*).
// Idiomatic Go has no equivalent "framework does everything" layer, so this
// file becomes an explicit, testable config loader driven by environment
// variables (twelve-factor style) with sane defaults matching the original
// values.
//
// The following original properties have NO direct Go equivalent and were
// intentionally dropped rather than faked:
//   - spring.datasource.driver-class-name: the Postgres driver is selected in
//     code by importing github.com/jackc/pgx/v5/stdlib (or lib/pq); there is no
//     runtime driver-class string.
//   - spring.jpa.hibernate.ddl-auto=update: schema management is not done by an
//     ORM at runtime. Use explicit SQL migrations (e.g. golang-migrate, goose)
//     instead. See DDLAuto field note below.
//   - spring.jpa.properties.hibernate.dialect: dialect is implied by the driver.
//
// MIGRATION_NOTE (dialect): the source targeted MySQL 8. Per the migration
// directive the target datastore is PostgreSQL, so the default DSN, port and
// driver assumptions here are PostgreSQL-shaped ($1 placeholders, RETURNING id,
// etc. live in the repository layer). Do NOT reintroduce the MySQL JDBC URL.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Default configuration values. These mirror the intent of the original
// application.properties while being PostgreSQL-appropriate.
const (
	defaultServerPort = 8082
	defaultDBHost     = "localhost"
	defaultDBPort     = 5432
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBName     = "barcode"
	defaultDBSSLMode  = "disable"
)

// Config holds the runtime configuration for the SmartContact application.
// It replaces the Spring Environment / @ConfigurationProperties abstraction
// with a plain, explicit struct.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on
	// (replaces server.port).
	ServerPort int

	// DBHost is the database host (parsed from the original JDBC URL).
	DBHost string
	// DBPort is the database port.
	DBPort int
	// DBUser is the database username (replaces spring.datasource.username).
	DBUser string
	// DBPassword is the database password (replaces spring.datasource.password).
	DBPassword string
	// DBName is the database/schema name (replaces the JDBC URL path segment).
	DBName string
	// DBSSLMode is the PostgreSQL sslmode connection parameter.
	DBSSLMode string

	// DDLAuto records the original spring.jpa.hibernate.ddl-auto intent.
	//
	// MIGRATION_NOTE: Go does not perform ORM-driven DDL. This field is retained
	// only as documentation/observability; it does NOT cause schema changes.
	// Run explicit migrations instead. Manual review recommended to ensure a
	// migration tool is wired up wherever the original relied on ddl-auto=update.
	DDLAuto string
}

// DSN builds a PostgreSQL connection string (keyword/value form) suitable for
// sql.Open with a pgx or lib/pq driver.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Addr returns the HTTP listen address in ":port" form, ready for http.Server.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// LoadConfig reads configuration from environment variables, falling back to
// the defaults derived from the original application.properties. It returns an
// error if a provided value cannot be parsed, so misconfiguration fails fast
// at startup rather than at first use.
//
// Recognised environment variables:
//
//	SERVER_PORT   -> ServerPort   (default 8082)
//	DB_HOST       -> DBHost       (default localhost)
//	DB_PORT       -> DBPort       (default 5432)
//	DB_USER       -> DBUser       (default postgres)
//	DB_PASSWORD   -> DBPassword   (default postgres)
//	DB_NAME       -> DBName       (default barcode)
//	DB_SSLMODE    -> DBSSLMode    (default disable)
//	DB_DDL_AUTO   -> DDLAuto      (default update; informational only)
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort: defaultServerPort,
		DBHost:     defaultDBHost,
		DBPort:     defaultDBPort,
		DBUser:     defaultDBUser,
		DBPassword: defaultDBPassword,
		DBName:     defaultDBName,
		DBSSLMode:  defaultDBSSLMode,
		DDLAuto:    "update",
	}

	var err error

	if cfg.ServerPort, err = intFromEnv("SERVER_PORT", cfg.ServerPort); err != nil {
		return nil, err
	}
	cfg.DBHost = stringFromEnv("DB_HOST", cfg.DBHost)
	if cfg.DBPort, err = intFromEnv("DB_PORT", cfg.DBPort); err != nil {
		return nil, err
	}
	cfg.DBUser = stringFromEnv("DB_USER", cfg.DBUser)
	cfg.DBPassword = stringFromEnv("DB_PASSWORD", cfg.DBPassword)
	cfg.DBName = stringFromEnv("DB_NAME", cfg.DBName)
	cfg.DBSSLMode = stringFromEnv("DB_SSLMODE", cfg.DBSSLMode)
	cfg.DDLAuto = stringFromEnv("DB_DDL_AUTO", cfg.DDLAuto)

	return cfg, nil
}

// stringFromEnv returns the trimmed value of the named environment variable, or
// the provided fallback when the variable is unset or empty.
func stringFromEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// intFromEnv parses the named environment variable as a base-10 integer,
// returning the fallback when unset/empty and an error when the value is
// present but not a valid integer.
func intFromEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s=%q: %w", key, raw, err)
	}
	return n, nil
}
