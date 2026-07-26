package resources

// Package resources holds application configuration for the SmartContact
// service. It corresponds to the original Spring Boot
// src/main/resources/application.properties file.
//
// MIGRATION_NOTE: Spring Boot's externalized configuration
// (application.properties) plus DataSource auto-configuration and Hibernate
// DDL/dialect handling have no idiomatic Go equivalent. There is no ORM
// auto-wiring; instead configuration is read explicitly from environment
// variables (with sane defaults) and used to construct a *sql.DB via the
// database/sql package. This keeps configuration explicit and testable.
//
// MIGRATION_NOTE: The original config targeted MySQL
// (jdbc:mysql://.../barcode, com.mysql.cj.jdbc.Driver, MySQL8Dialect). Per the
// migration target, the Go implementation uses PostgreSQL instead. The DSN,
// driver, and default port (5432) reflect that decision — this is an
// intentional dialect switch, not a mirror of the source. Review the default
// credentials and connection parameters before deploying.
//
// MIGRATION_NOTE: Hibernate's spring.jpa.hibernate.ddl-auto=update (automatic
// schema migration) has NO safe Go equivalent and is deliberately dropped.
// Schema management should be handled by an explicit migration tool such as
// golang-migrate or goose. See RunMigrations placeholder guidance below.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Default configuration values. These mirror the intent of the original
// application.properties but target PostgreSQL rather than MySQL.
const (
	// DefaultServerPort is the HTTP port the service listens on.
	// Migrated from server.port=8082.
	DefaultServerPort = 8082

	// DefaultDBHost is the database host. Migrated from the host portion of
	// jdbc:mysql://localhost:3306/barcode.
	DefaultDBHost = "localhost"

	// DefaultDBPort is the PostgreSQL port (the source used MySQL's 3306; the
	// PostgreSQL default 5432 is used per the target dialect).
	DefaultDBPort = 5432

	// DefaultDBName is the database/schema name. Migrated from the "barcode"
	// path segment of the JDBC URL.
	DefaultDBName = "barcode"

	// DefaultDBUser is the database user. Migrated from
	// spring.datasource.username=root.
	//
	// MIGRATION_NOTE: "root" is a MySQL convention; PostgreSQL typically uses
	// "postgres". The source value is preserved as a default, but review this.
	DefaultDBUser = "postgres"

	// DefaultDBPassword is the database password. Migrated from
	// spring.datasource.password=root. Do NOT rely on this default in
	// production — supply DB_PASSWORD via the environment or a secret store.
	DefaultDBPassword = "root"

	// DefaultDBSSLMode controls PostgreSQL TLS negotiation. "disable" matches
	// the local-development posture of the original config.
	DefaultDBSSLMode = "disable"
)

// Config holds the fully resolved application configuration. It replaces the
// Spring Boot property binding performed against application.properties.
type Config struct {
	// ServerPort is the HTTP listen port.
	ServerPort int

	// DBHost is the database server host.
	DBHost string
	// DBPort is the database server port.
	DBPort int
	// DBName is the database name.
	DBName string
	// DBUser is the database user.
	DBUser string
	// DBPassword is the database password.
	DBPassword string
	// DBSSLMode is the PostgreSQL sslmode connection parameter.
	DBSSLMode string
}

// LoadConfig builds a Config from environment variables, falling back to the
// defaults derived from the original application.properties. It never panics;
// invalid numeric values are reported as errors.
//
// Recognised environment variables:
//
//	SERVER_PORT, DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
func LoadConfig() (*Config, error) {
	serverPort, err := envInt("SERVER_PORT", DefaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("loading SERVER_PORT: %w", err)
	}

	dbPort, err := envInt("DB_PORT", DefaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("loading DB_PORT: %w", err)
	}

	return &Config{
		ServerPort: serverPort,
		DBHost:     envString("DB_HOST", DefaultDBHost),
		DBPort:     dbPort,
		DBName:     envString("DB_NAME", DefaultDBName),
		DBUser:     envString("DB_USER", DefaultDBUser),
		DBPassword: envString("DB_PASSWORD", DefaultDBPassword),
		DBSSLMode:  envString("DB_SSLMODE", DefaultDBSSLMode),
	}, nil
}

// DSN returns a PostgreSQL data source name suitable for sql.Open with the
// lib/pq or pgx stdlib driver.
//
// MIGRATION_NOTE: This replaces spring.datasource.url. The original JDBC URL
// pointed at MySQL; the DSN here is PostgreSQL-shaped per the target dialect.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// ServerAddr returns the address string (":port") for http.Server.Addr.
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// OpenDB opens and verifies a database connection pool using the resolved
// configuration. The context bounds the initial connectivity check (PingContext).
//
// MIGRATION_NOTE: This replaces Spring Boot's DataSource auto-configuration.
// The caller is responsible for closing the returned *sql.DB. The driverName
// ("postgres") assumes a lib/pq-compatible driver is registered via a blank
// import in main; adjust to "pgx" if using jackc/pgx's stdlib driver.
func (c *Config) OpenDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Reasonable pool defaults; Spring Boot/HikariCP applied its own.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

// envString returns the value of the environment variable named key, or def if
// the variable is unset or empty.
func envString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// envInt returns the integer value of the environment variable named key, or
// def if the variable is unset or empty. It returns an error if the value is
// present but not a valid integer.
func envInt(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("parsing %s=%q as int: %w", key, v, err)
	}
	return n, nil
}
