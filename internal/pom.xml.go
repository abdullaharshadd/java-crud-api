// Package internal documents the migration of the Maven Project Object
// Model (pom.xml) to the Go module system.
//
// MIGRATION_NOTE: pom.xml is a Maven build descriptor, not source code. It
// has no runtime behavior to translate; there is no idiomatic Go "file"
// equivalent of a POM. Go's build/dependency metadata lives in go.mod and
// go.sum at the repository root, driven by the `go` toolchain rather than a
// declarative XML descriptor consumed by a plugin-based build engine like
// Maven.
//
// This file exists only to record, in one place, how each Maven coordinate
// from the original SmartContact POM maps onto the Go ecosystem, and to
// point at the real go.mod that must be committed at the repository root.
// It declares no runtime types or functions, so it cannot collide with any
// already-migrated symbol.
//
// ---------------------------------------------------------------------------
// Project identity
// ---------------------------------------------------------------------------
//   Maven                              Go
//   groupId  com.smartContact         module path segment
//   artifactId SmartContact           repository / module name
//   version  0.0.1-SNAPSHOT           git tag (e.g. v0.0.1) once released
//   java.version 17                   go directive in go.mod (e.g. go 1.22)
//
// ---------------------------------------------------------------------------
// Dependency mapping (Maven starter -> Go library)
// ---------------------------------------------------------------------------
//   spring-boot-starter-web        -> github.com/go-chi/chi/v5 (routing) plus
//                                     the standard library net/http server.
//                                     The embedded Tomcat container is
//                                     replaced by net/http's own server,
//                                     wired in smartcontactapplication.go.
//   spring-boot-starter-data-jpa   -> database/sql + github.com/lib/pq
//                                     (PostgreSQL driver). Hibernate ORM and
//                                     its ddl-auto schema management become
//                                     explicit SQL: see EnsureUserSchema in
//                                     the composition root and the repository
//                                     layer (internal/smartcontact/repository).
//   mysql-connector-j              -> DROPPED. The target database is
//                                     PostgreSQL, so the MySQL JDBC driver is
//                                     replaced by github.com/lib/pq. Do not
//                                     reintroduce a MySQL driver; the query
//                                     layer uses $1,$2 placeholders and
//                                     RETURNING id, which are Postgres idioms.
//   spring-boot-starter-validation -> github.com/go-playground/validator/v10
//                                     (or hand-written checks). Bean
//                                     Validation annotations have no Go
//                                     equivalent; validation is explicit.
//   spring-boot-starter-thymeleaf  -> the standard library html/template
//                                     package, if server-side HTML rendering
//                                     is actually required. REQUIRES MANUAL
//                                     REVIEW: the migrated handlers currently
//                                     serve JSON, so this may be unnecessary.
//   spring-boot-devtools           -> DROPPED. Live-reload/hot-restart has no
//                                     stdlib analogue; use an external watcher
//                                     such as github.com/cosmtrek/air during
//                                     development if desired.
//   lombok                         -> DROPPED. Compile-time boilerplate
//                                     generation is unnecessary in Go; structs
//                                     are written out explicitly.
//   spring-boot-starter-test       -> github.com/stretchr/testify plus the
//                                     standard library testing package. See
//                                     the table-driven tests under
//                                     internal/smartcontact/service/service_test.
//
// ---------------------------------------------------------------------------
// Build plugin mapping
// ---------------------------------------------------------------------------
//   spring-boot-maven-plugin       -> `go build ./cmd/...` produces a single
//                                     statically-linked binary, so no
//                                     repackaging plugin is needed. The Lombok
//                                     <excludes> block is moot because Lombok
//                                     is dropped entirely.
//
// ---------------------------------------------------------------------------
// Suggested go.mod (commit this at the repository root, NOT here)
// ---------------------------------------------------------------------------
//   module github.com/smartcontact/smartcontact
//
//   go 1.22
//
//   require (
//           github.com/go-chi/chi/v5 v5.0.12
//           github.com/lib/pq v1.10.9
//           github.com/go-playground/validator/v10 v10.19.0
//           github.com/stretchr/testify v1.9.0
//   )
package internal
