// Package internal documents the migration of the Maven Project Object
// Model (pom.xml) for the SmartContact Spring Boot application.
//
// MIGRATION_NOTE: pom.xml is a BUILD DESCRIPTOR, not application source
// code. It carries no runtime business logic, HTTP routes, or Spring
// beans. There is nothing to execute here. In a Go project the entire
// Maven build/dependency layer is replaced by the Go toolchain and
// go.mod. This file exists solely to record the migration mapping so a
// human reviewer can reconstruct the equivalent Go dependency set.
//
// ---------------------------------------------------------------------
// Project metadata
// ---------------------------------------------------------------------
//
//   Maven groupId/artifactId/version : com.smartContact / SmartContact / 0.0.1-SNAPSHOT
//   Java runtime                     : 17
//   Spring Boot                      : 2.7.14  (javax.* namespace, NOT jakarta.*)
//
// Go equivalent: declare the module path and Go version in go.mod, e.g.
//
//   module github.com/smartcontact/smartcontact
//   go 1.22
//
// ---------------------------------------------------------------------
// Dependency mapping (Maven -> Go)
// ---------------------------------------------------------------------
//
//   spring-boot-starter-web        -> HTTP router/framework.
//                                     Recommended: github.com/go-chi/chi/v5
//                                     (net/http compatible, idiomatic middleware).
//
//   spring-boot-starter-thymeleaf  -> Server-side HTML templating.
//                                     Use the standard library html/template.
//                                     This app is server-rendered MVC, NOT a
//                                     pure JSON API — templates must be ported
//                                     from src/main/resources/templates/*.html.
//
//   spring-boot-starter-data-jpa   -> ORM / data access (Hibernate under the hood).
//                                     Recommended: gorm.io/gorm with
//                                     gorm.io/driver/mysql, OR raw database/sql
//                                     with github.com/go-sql-driver/mysql for a
//                                     leaner, more explicit approach.
//
//   mysql-connector-j              -> github.com/go-sql-driver/mysql (driver).
//                                     HikariCP pooling is replaced by the
//                                     database/sql pool: configure via
//                                     db.SetMaxOpenConns / SetMaxIdleConns /
//                                     SetConnMaxLifetime after sql.Open.
//
//   spring-boot-starter-validation -> github.com/go-playground/validator/v10.
//                                     Map bean-validation annotations
//                                     (@NotNull, @Email, @Size, ...) to struct
//                                     tags, e.g. `validate:"required,email"`.
//
//   lombok                         -> No equivalent needed. Lombok is a
//                                     compile-time-only JVM code generator for
//                                     getters/setters/builders/constructors.
//                                     In Go these are plain exported struct
//                                     fields and explicit NewXxx constructors.
//
//   spring-boot-devtools           -> No runtime equivalent. Dev-only hot
//                                     reload. Use a Go file watcher such as
//                                     github.com/air-verse/air locally.
//
//   spring-boot-starter-test       -> Standard library testing + optionally
//                                     github.com/stretchr/testify. Write
//                                     table-driven tests per migrated package.
//
// ---------------------------------------------------------------------
// Build/packaging mapping
// ---------------------------------------------------------------------
//
//   spring-boot-maven-plugin (fat/uber JAR) -> `go build ./cmd/smartcontact`
//     produces a single self-contained static binary; no fat-JAR tooling
//     required. Embed templates/static assets with the //go:embed directive
//     so the binary is fully self-contained.
//
//   Lombok exclude in the plugin config     -> Not applicable (no Lombok).
//
// ---------------------------------------------------------------------
// REQUIRES MANUAL REVIEW
// ---------------------------------------------------------------------
//
//   1. javax.* vs jakarta.* namespace choice is irrelevant in Go, but any
//      persisted schema semantics from JPA entities must be reproduced
//      exactly in the Go models/migrations.
//   2. No HTTP routes are defined in pom.xml — routes live in the Spring
//      @Controller classes and must be registered when those files are
//      migrated (e.g. with chi.Router.Get/Post).
//   3. Confirm MySQL connection-pool parameters match the HikariCP defaults
//      previously in effect.
package internal
