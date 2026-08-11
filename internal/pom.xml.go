// Package internal documents the migration of the source project's Maven POM
// (pom.xml) into the Go module ecosystem.
//
// MIGRATION_NOTE: pom.xml is a Maven build-configuration artifact. It has NO
// runtime behavior and therefore no direct executable Go equivalent. Maven's
// role — declaring dependencies, pinning versions, and driving the build — is
// handled in Go by the `go` toolchain and the module manifest (go.mod).
//
// This file exists purely to preserve the migration mapping for humans auditing
// the port. Nothing here is executed by the application. It is safe to delete
// once the migration has been reviewed and go.mod is authoritative.
//
// Dependency mapping (Spring Boot starter -> Go equivalent):
//
//	spring-boot-starter-web        -> net/http + a router (github.com/go-chi/chi/v5)
//	spring-boot-starter-data-jpa   -> database/sql + github.com/jmoiron/sqlx
//	mysql-connector-j              -> DROPPED. Target DB is PostgreSQL; use
//	                                  github.com/jackc/pgx/v5 (pgx stdlib driver).
//	                                  Do NOT port the MySQL driver — the target
//	                                  dialect is Postgres ($1 placeholders,
//	                                  RETURNING id, SERIAL/IDENTITY PKs).
//	spring-boot-starter-validation -> github.com/go-playground/validator/v10
//	spring-boot-starter-thymeleaf  -> html/template (stdlib) for server-side HTML
//	spring-boot-devtools           -> No equivalent; use `air`/`go run` for reload
//	lombok                         -> No equivalent; Go has no annotation
//	                                  processing. Write structs/constructors
//	                                  explicitly (NewXxx functions).
//	spring-boot-starter-test       -> testing (stdlib) + github.com/stretchr/testify
//	spring-boot-maven-plugin       -> `go build` produces a self-contained
//	                                  static binary; no fat-JAR packaging needed.
//
// Java version: 17 -> Go toolchain version is declared via the `go` directive
// in go.mod (e.g. `go 1.22`).
//
// Config already migrated in internal/config/config.go reads DATABASE_URL,
// PORT, and JWT_SECRET from the environment via viper.
package internal
