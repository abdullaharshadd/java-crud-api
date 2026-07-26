// Package resources holds the application's externalized configuration.
//
// MIGRATION_NOTE: The Java source was src/main/resources/application.properties,
// a Spring Boot externalized-configuration file. Spring Boot auto-binds these
// keys into the framework's DataSource, JPA/Hibernate, and embedded-server
// beans via convention-over-configuration. Go has no equivalent auto-binding
// mechanism: configuration is loaded explicitly and passed to constructors via
// dependency injection.
//
// The idiomatic translation is a typed Config struct with a loader that reads
// from environment variables (12-factor style), falling back to the original
// property values as defaults. This keeps the concrete connection settings out
// of the compiled binary while preserving the exact behaviour of the original
// file when no environment overrides are present.
//
// Notable per-key mapping decisions:
//
//   - server.port                        -> Config.ServerPort (used to build the
//     net/http listen address, e.g. ":8082").
//   - spring.datasource.url (JDBC)       -> converted to a Go database/sql DSN.
//     JDBC URLs (jdbc:mysql://host:port/db) are NOT valid Go MySQL DSNs. The
//     go-sql-driver/mysql format is "user:pass@tcp(host:port)/db". BuildDSN()
//     performs this conversion explicitly.
//   - spring.datasource.driver-class-name -> DROPPED. The Go driver is selected
//     at compile time via a blank import of github.com/go-sql-driver/mysql;
//     there is no runtime driver-class string.
//   - spring.jpa.hibernate.ddl-auto=update -> NOT AUTOMATICALLY MIGRATED. Go has
//     no ORM performing automatic schema migration. This must be handled by an
//     explicit migration tool (e.g. golang-migrate, goose). See notes.
//   - spring.jpa.properties.hibernate.dialect -> DROPPED. Dialect selection is
//     implicit in the chosen database/sql driver; there is no dialect concept.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values transcribed verbatim from the original
// application.properties. They are used when the corresponding environment
// variable is unset, preserving the original file's behaviour.
const (
	// DefaultServerPort mirrors server.port=8082.
	DefaultServerPort = 8082
	// DefaultDBHost is derived from spring.datasource.url host component.
	DefaultDBHost = "localhost"
	// DefaultDBPort is derived from spring.datasource.url port component.
	DefaultDBPort = 3306
	// DefaultDBName mirrors the schema in spring.datasource.url (barcode).
	DefaultDBName = "barcode"
	// DefaultDBUser mirrors spring.datasource.username=root.
	DefaultDBUser = "root"
	// DefaultDBPassword mirrors spring.datasource.password=root.
	//
	// MIGRATION_NOTE: A hardcoded credential is a security smell carried over
	// from the source file. It is retained only as a development default and
	// should be overridden by the DB_PASSWORD environment variable in any real
	// deployment.
	DefaultDBPassword = "root"
)

// Config holds the application's runtime configuration. It is the idiomatic
// replacement for Spring Boot's auto-bound configuration properties.
type Config struct {
	// ServerPort is the TCP port the HTTP server listens on.
	ServerPort int
	// DBHost is the database server hostname.
	DBHost string
	// DBPort is the database server TCP port.
	DBPort int
	// DBName is the target schema/database name.
	DBName string
	// DBUser is the database username.
	DBUser string
	// DBPassword is the database password.
	DBPassword string
}

// NewConfig builds a Config from the process environment, falling back to the
// defaults transcribed from application.properties. It returns an error if any
// numeric environment override cannot be parsed.
//
// Recognised environment variables:
//
//	SERVER_PORT, DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
func NewConfig() (*Config, error) {
	cfg := &Config{
		ServerPort: DefaultServerPort,
		DBHost:     DefaultDBHost,
		DBPort:     DefaultDBPort,
		DBName:     DefaultDBName,
		DBUser:     DefaultDBUser,
		DBPassword: DefaultDBPassword,
	}

	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("resources: invalid SERVER_PORT %q: %w", v, err)
		}
		cfg.ServerPort = port
	}

	if v, ok := os.LookupEnv("DB_HOST"); ok {
		cfg.DBHost = v
	}

	if v, ok := os.LookupEnv("DB_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("resources: invalid DB_PORT %q: %w", v, err)
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

	return cfg, nil
}

// BuildDSN returns a data source name suitable for sql.Open with the
// github.com/go-sql-driver/mysql driver.
//
// MIGRATION_NOTE: The source used a JDBC URL
// (jdbc:mysql://localhost:3306/barcode). That format is not understood by the
// Go MySQL driver, which expects "user:password@tcp(host:port)/dbname". The
// parseTime=true and multiStatements parameters are commonly required; only
// parseTime is enabled here to keep behaviour conservative and predictable.
func (c *Config) BuildDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// ListenAddr returns the address string for http.Server.Addr, e.g. ":8082".
func (c *Config) ListenAddr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}
