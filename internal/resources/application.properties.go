// Package resources holds the application's externalized configuration.
//
// MIGRATION_NOTE: The Java source was src/main/resources/application.properties,
// a Spring Boot externalized configuration file. Spring Boot auto-configured an
// embedded servlet container, a MySQL datasource, and a Hibernate/JPA
// EntityManager purely from these properties at runtime via classpath scanning
// and auto-configuration.
//
// Go has no equivalent runtime auto-configuration mechanism, so this file
// translates the properties into an explicit, strongly-typed Config struct plus
// a constructor that reads values from the environment (with the .properties
// values as documented defaults). The composition root (cmd/server/main.go)
// is expected to call NewConfig and pass the resulting *sql.DB / address into
// the wiring in internal/smartcontact/smartcontactapplication.go.
//
// IMPORTANT DIALECT CHANGE — MANUAL REVIEW REQUIRED:
// The source targeted MySQL (jdbc:mysql, com.mysql.cj.jdbc.Driver,
// MySQL8Dialect). The migration target is PostgreSQL. The defaults below have
// been translated to a PostgreSQL DSN and the lib/pq / pgx driver name.
// Repository/DAO code must use $1, $2 ... placeholders and RETURNING id for
// inserts — do NOT carry over MySQL's `?` placeholders or LastInsertId().
//
// MIGRATION_NOTE: spring.jpa.hibernate.ddl-auto=update performed implicit schema
// migration at startup. Go has no ORM doing this automatically; schema
// management must be handled explicitly (e.g. golang-migrate, goose, or a
// checked-in SQL migration). AutoMigrate below is intentionally a documented
// no-op that returns an error directing the operator to a real migration tool,
// so the implicit-migration behavior is not silently lost.
package resources

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Default configuration values, translated from the original
// application.properties. The MySQL-specific values have deliberately been
// re-expressed for the PostgreSQL target (see package doc).
const (
	// DefaultServerPort mirrors server.port=8082 from the source properties.
	DefaultServerPort = 8082

	// DefaultDBHost is the PostgreSQL host (source used localhost for MySQL).
	DefaultDBHost = "localhost"

	// DefaultDBPort is the PostgreSQL port. NOTE: the source used MySQL's 3306;
	// PostgreSQL's default is 5432.
	DefaultDBPort = 5432

	// DefaultDBName mirrors the "barcode" database name from the source.
	DefaultDBName = "barcode"

	// DefaultDBUser mirrors spring.datasource.username=root.
	DefaultDBUser = "root"

	// DefaultDBPassword mirrors spring.datasource.password=root.
	//
	// MANUAL REVIEW: committing credentials is a security anti-pattern carried
	// over from the source. Override via PGPASSWORD/DB_PASSWORD in real
	// environments and remove this default before production use.
	DefaultDBPassword = "root"

	// DefaultSSLMode is the PostgreSQL sslmode. "disable" matches the plaintext
	// local-dev posture implied by the source's jdbc:mysql URL.
	DefaultSSLMode = "disable"

	// DriverName is the database/sql driver name for PostgreSQL.
	// MIGRATION_NOTE: replaces com.mysql.cj.jdbc.Driver. Register the matching
	// driver (e.g. _ "github.com/lib/pq") in the composition root.
	DriverName = "postgres"
)

// Config holds the fully-resolved application configuration. It replaces the
// key/value pairs Spring Boot would have bound from application.properties.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on (source: server.port).
	ServerPort int

	// DBHost is the database server hostname.
	DBHost string
	// DBPort is the database server port.
	DBPort int
	// DBName is the database/schema name.
	DBName string
	// DBUser is the database username.
	DBUser string
	// DBPassword is the database password.
	DBPassword string
	// SSLMode is the PostgreSQL sslmode connection parameter.
	SSLMode string

	// AutoMigrateSchema mirrors spring.jpa.hibernate.ddl-auto=update: when true
	// the operator has opted into automatic schema management. See AutoMigrate.
	AutoMigrateSchema bool
}

// NewConfig builds a Config from environment variables, falling back to the
// defaults translated from the original application.properties. It returns an
// error if any provided value is malformed. This replaces Spring Boot's
// property binding and @Value injection.
func NewConfig() (*Config, error) {
	port, err := intFromEnv("SERVER_PORT", DefaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("resources: invalid SERVER_PORT: %w", err)
	}

	dbPort, err := intFromEnv("DB_PORT", DefaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("resources: invalid DB_PORT: %w", err)
	}

	autoMigrate, err := boolFromEnv("DB_AUTO_MIGRATE", true) // ddl-auto=update -> true
	if err != nil {
		return nil, fmt.Errorf("resources: invalid DB_AUTO_MIGRATE: %w", err)
	}

	return &Config{
		ServerPort:        port,
		DBHost:            stringFromEnv("DB_HOST", DefaultDBHost),
		DBPort:            dbPort,
		DBName:            stringFromEnv("DB_NAME", DefaultDBName),
		DBUser:            stringFromEnv("DB_USER", DefaultDBUser),
		DBPassword:        stringFromEnv("DB_PASSWORD", DefaultDBPassword),
		SSLMode:           stringFromEnv("DB_SSLMODE", DefaultSSLMode),
		AutoMigrateSchema: autoMigrate,
	}, nil
}

// ServerAddr returns the address suitable for http.Server.Addr, e.g. ":8082".
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// DSN returns a PostgreSQL data source name for database/sql.Open.
//
// MIGRATION_NOTE: replaces spring.datasource.url
// (jdbc:mysql://localhost:3306/barcode). The format is the lib/pq keyword DSN.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.SSLMode,
	)
}

// OpenDB opens and verifies a connection pool using the resolved configuration.
// It replaces Spring Boot's datasource auto-configuration. The caller owns the
// returned *sql.DB and is responsible for closing it.
func (c *Config) OpenDB() (*sql.DB, error) {
	db, err := sql.Open(DriverName, c.DSN())
	if err != nil {
		return nil, fmt.Errorf("resources: opening database: %w", err)
	}

	// Sensible connection-pool defaults; Spring/Hikari applied similar tuning
	// implicitly. Adjust for your workload.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("resources: pinging database: %w", err)
	}
	return db, nil
}

// ErrManualMigration signals that automatic schema migration is not supported.
var ErrManualMigration = errors.New(
	"resources: automatic schema migration (Hibernate ddl-auto=update) has no " +
		"Go equivalent; run explicit migrations (golang-migrate/goose) instead",
)

// AutoMigrate is the explicit counterpart to spring.jpa.hibernate.ddl-auto=update.
//
// MIGRATION_NOTE: Hibernate silently altered the live schema at startup. That
// behavior is intentionally NOT reproduced: silent DDL against a production
// database is dangerous. When AutoMigrateSchema is enabled this returns
// ErrManualMigration so the operator is forced to wire in a real migration
// tool; when disabled it is a no-op. This must be reviewed and replaced with a
// real migration step.
func (c *Config) AutoMigrate(_ *sql.DB) error {
	if !c.AutoMigrateSchema {
		return nil
	}
	return ErrManualMigration
}

// stringFromEnv returns the environment value for key, or def if unset/empty.
func stringFromEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// intFromEnv returns the integer environment value for key, or def if unset.
func intFromEnv(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("parsing %q=%q as int: %w", key, v, err)
	}
	return n, nil
}

// boolFromEnv returns the boolean environment value for key, or def if unset.
func boolFromEnv(key string, def bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("parsing %q=%q as bool: %w", key, v, err)
	}
	return b, nil
}
