// Package resources holds application configuration that was previously
// expressed as Spring Boot externalized configuration.
//
// MIGRATION_NOTE: The Java source was src/main/resources/application.properties,
// a Spring Boot property file. Spring Boot read this at startup to:
//   - configure the embedded server port (server.port),
//   - auto-configure a DataSource (spring.datasource.*),
//   - auto-configure Hibernate/JPA schema management and dialect
//     (spring.jpa.hibernate.ddl-auto, spring.jpa.properties.hibernate.dialect).
//
// Go has no equivalent auto-configuration magic and no ORM performing implicit
// DDL generation. The idiomatic replacement is an explicit, typed Config struct
// that is populated from environment variables (with sane defaults) and used to
// build a *sql.DB and an HTTP server address. This keeps configuration explicit
// and testable rather than relying on classpath scanning.
//
// IMPORTANT DIALECT CHANGE: the original file targeted MySQL
// (jdbc:mysql, com.mysql.cj.jdbc.Driver, MySQL8Dialect). Per the migration
// target, the datasource is now PostgreSQL. The default DSN below reflects that.
// Register a PostgreSQL driver (e.g. github.com/jackc/pgx/v5/stdlib or
// github.com/lib/pq) in the composition root before calling Open.
//
// MANUAL REVIEW:
//   - spring.jpa.hibernate.ddl-auto=update auto-migrated the schema at startup.
//     Go has no ORM doing this; run explicit migrations (e.g. golang-migrate,
//     goose, or hand-written SQL) instead. RunMigrations here is a deliberate
//     no-op placeholder documenting that requirement.
//   - The plaintext credentials (root/root) from the source are only defaults
//     for local development; supply real values via environment variables in
//     any non-local environment.
package resources

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Default configuration values mirroring the original application.properties.
// The datasource defaults have been translated from MySQL to PostgreSQL.
const (
	defaultServerPort = 8082
	defaultDBHost     = "localhost"
	defaultDBPort     = 5432
	defaultDBName     = "barcode"
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBSSLMode  = "disable"
)

// Config holds the application configuration previously defined in
// application.properties. Values are resolved from environment variables with
// defaults applied when a variable is unset.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on (was server.port).
	ServerPort int

	// DBHost is the database server host.
	DBHost string
	// DBPort is the database server port.
	DBPort int
	// DBName is the database name (was the "barcode" schema in the JDBC URL).
	DBName string
	// DBUser is the database user (was spring.datasource.username).
	DBUser string
	// DBPassword is the database password (was spring.datasource.password).
	DBPassword string
	// DBSSLMode is the PostgreSQL sslmode connection parameter.
	DBSSLMode string
}

// LoadConfig builds a Config from environment variables, falling back to the
// defaults derived from the original application.properties. It returns an
// error if a numeric environment variable is set but cannot be parsed.
func LoadConfig() (*Config, error) {
	serverPort, err := intFromEnv("SERVER_PORT", defaultServerPort)
	if err != nil {
		return nil, fmt.Errorf("parsing SERVER_PORT: %w", err)
	}

	dbPort, err := intFromEnv("DB_PORT", defaultDBPort)
	if err != nil {
		return nil, fmt.Errorf("parsing DB_PORT: %w", err)
	}

	return &Config{
		ServerPort: serverPort,
		DBHost:     stringFromEnv("DB_HOST", defaultDBHost),
		DBPort:     dbPort,
		DBName:     stringFromEnv("DB_NAME", defaultDBName),
		DBUser:     stringFromEnv("DB_USER", defaultDBUser),
		DBPassword: stringFromEnv("DB_PASSWORD", defaultDBPassword),
		DBSSLMode:  stringFromEnv("DB_SSLMODE", defaultDBSSLMode),
	}, nil
}

// ServerAddr returns the address the HTTP server should bind to, suitable for
// passing to http.Server.Addr (e.g. ":8082").
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// DSN returns a PostgreSQL connection string suitable for sql.Open with a
// PostgreSQL driver.
//
// MIGRATION_NOTE: replaces the MySQL JDBC URL/driver from the source. The
// caller must register a PostgreSQL driver before opening a connection.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// OpenDB opens and verifies a database connection using the given driver name
// (for example "pgx" or "postgres") and this configuration.
//
// MIGRATION_NOTE: Spring Boot auto-configured the DataSource from the
// spring.datasource.* properties. Here we open it explicitly and ping to fail
// fast, honoring the migration note that the DB connection must be established
// before the repository layer is wired.
func (c *Config) OpenDB(ctx context.Context, driverName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, c.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening database (driver=%q): %w", driverName, err)
	}

	// Conservative, production-sane pool defaults; Spring Boot/HikariCP applied
	// similar implicit tuning.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		// Best-effort close; the ping error is the meaningful one to surface.
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

// RunMigrations is a placeholder for schema management.
//
// MIGRATION_NOTE: The source set spring.jpa.hibernate.ddl-auto=update, which
// let Hibernate mutate the schema automatically at startup. Go has no ORM
// performing that step. This function intentionally does nothing and exists to
// document that explicit migrations (golang-migrate, goose, or hand-written
// SQL executed against db) must replace Hibernate's auto-DDL. Wire a real
// migration tool here or remove this function once migrations live elsewhere.
func RunMigrations(_ context.Context, _ *sql.DB) error {
	return nil
}

// stringFromEnv returns the value of the named environment variable, or def if
// it is unset or empty.
func stringFromEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// intFromEnv returns the integer value of the named environment variable, or
// def if it is unset or empty. It returns an error if the value is set but not
// a valid integer.
func intFromEnv(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", v, err)
	}
	return n, nil
}
