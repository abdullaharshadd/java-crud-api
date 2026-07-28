// Package internal previously corresponded to the Maven Project Object
// Model (pom.xml) for the SmartContact Spring Boot application.
//
// MIGRATION_NOTE: pom.xml is a build/dependency-management descriptor for the
// Maven + JVM ecosystem. It contains NO business logic and therefore has no
// direct Go source-code counterpart. In an idiomatic Go project the same
// responsibilities are handled by the Go module system and the standard
// toolchain. This file documents the intended translation so that the mapping
// from the original Maven configuration is auditable.
//
// The equivalent Go project setup is described below rather than executed as
// code, because dependency declaration belongs in go.mod, not in a .go file.
//
// ----------------------------------------------------------------------------
// Maven parent / packaging  ->  Go module + `go build`
// ----------------------------------------------------------------------------
// The Spring Boot parent BOM (spring-boot-starter-parent 2.7.14) pinned
// dependency versions and produced a fat JAR via spring-boot-maven-plugin.
// In Go, `go build ./cmd/server` already produces a single statically-linked
// binary, so no fat-JAR plugin is needed.
//
// ----------------------------------------------------------------------------
// Java 17 target  ->  go directive in go.mod
// ----------------------------------------------------------------------------
// <java.version>17</java.version> maps to a `go 1.23` (or similar) directive
// plus an optional `toolchain` directive in go.mod.
//
// ----------------------------------------------------------------------------
// Dependency mapping (Maven artifact  ->  Go module)
// ----------------------------------------------------------------------------
//   spring-boot-starter-web        -> net/http (stdlib) + github.com/go-chi/chi/v5
//   spring-boot-starter-thymeleaf  -> html/template (stdlib) for server-side
//                                     rendering (MVC, not REST-only)
//   spring-boot-starter-data-jpa   -> github.com/jmoiron/sqlx over
//                                     database/sql; schema managed with
//                                     github.com/golang-migrate/migrate/v4
//                                     (replaces Hibernate ddl-auto)
//   spring-boot-starter-validation -> github.com/go-playground/validator/v10
//   mysql-connector-j              -> github.com/jackc/pgx/v5 (PostgreSQL is the
//                                     TARGET database; the source MySQL driver
//                                     is intentionally NOT mirrored)
//   lombok                         -> no equivalent; Go generates no boilerplate
//                                     at compile time. Struct fields, getters
//                                     and constructors are written explicitly
//                                     (or via `go generate` if desired).
//   spring-boot-devtools           -> no equivalent; use a live-reload tool such
//                                     as github.com/air-verse/air during
//                                     development (dev-only, not vendored).
//   spring-boot-starter-test       -> testing (stdlib) +
//                                     github.com/stretchr/testify
//
// ----------------------------------------------------------------------------
// Spring component scanning / DI  ->  explicit constructor wiring
// ----------------------------------------------------------------------------
// Spring's classpath component scanning is removed entirely. Dependencies are
// wired explicitly via NewXxx constructors in cmd/server/main.go (see
// buildRouter), optionally with github.com/google/wire for compile-time DI.
//
// ----------------------------------------------------------------------------
// Suggested go.mod (create at the repository root, not in this file):
//
//   module migrated-app/smartcontact
//
//   go 1.23
//
//   require (
//       github.com/go-chi/chi/v5             v5.0.12
//       github.com/jmoiron/sqlx              v1.4.0
//       github.com/jackc/pgx/v5              v5.6.0
//       github.com/golang-migrate/migrate/v4 v4.17.1
//       github.com/go-playground/validator/v10 v10.22.0
//       github.com/rs/zerolog                v1.33.0
//       github.com/spf13/viper               v1.19.0
//       github.com/stretchr/testify          v1.9.0
//   )
//
// ----------------------------------------------------------------------------
//
// There is no exported API to declare here: build configuration is not runtime
// code. This doc comment is the complete, intentional translation of pom.xml.
package internal
