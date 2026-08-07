// Package internal documents the project's build and dependency configuration.
//
// MIGRATION_NOTE: The source file was pom.xml — a Maven Project Object Model
// used by Spring Boot to declare dependencies, the Java language level, and
// build/packaging plugins. It contains no business logic and no executable
// code; it is pure build metadata resolved at build time by Maven.
//
// Go has no direct equivalent to a POM. The Go toolchain uses:
//
//   - go.mod / go.sum      for module identity, dependency declarations, and
//                          cryptographic version pinning (replaces <groupId>,
//                          <artifactId>, <version>, and <dependencies>).
//   - the `go` directive    in go.mod for the language version (replaces the
//                          <java.version>17</java.version> property).
//   - `go build` / `go install`
//                          for producing a native, statically-linked binary
//                          (replaces the spring-boot-maven-plugin's executable
//                          JAR repackaging goal).
//
// Because none of this belongs in a .go source file, this file is
// documentation only. Below is the mapping a human should apply when creating
// the real go.mod. There is intentionally no runtime code here.
//
// ---------------------------------------------------------------------------
// Maven coordinates -> Go module
// ---------------------------------------------------------------------------
//   <groupId>com.smartContact</groupId>
//   <artifactId>SmartContact</artifactId>
//   <version>0.0.1-SNAPSHOT</version>
//
//   => module github.com/smartcontact/smartcontact   (adjust to your VCS host)
//      go 1.21                                        (or newer; >= the Java 17
//                                                       era baseline)
//
// ---------------------------------------------------------------------------
// Dependency mapping (Spring Boot starters have no 1:1 Go analogue; they are
// broken down into the focused libraries the code actually imports)
// ---------------------------------------------------------------------------
//   spring-boot-starter-web        -> net/http (stdlib) + a router such as
//                                     github.com/go-chi/chi/v5
//   spring-boot-starter-data-jpa   -> database/sql (stdlib). No ORM is assumed;
//                                     the already-migrated resources package
//                                     (internal/resources/application.properties.go)
//                                     exposes OpenDB / AutoMigrate over *sql.DB.
//   mysql-connector-j (runtime)    -> DIALECT CHANGE: the target datastore for
//                                     this migration is PostgreSQL, NOT MySQL.
//                                     Use github.com/jackc/pgx/v5/stdlib (or
//                                     github.com/lib/pq) and $1,$2 placeholders
//                                     with RETURNING id on INSERT. Do NOT carry
//                                     the MySQL driver forward.
//   spring-boot-starter-validation -> hand-written validation, or
//                                     github.com/go-playground/validator/v10.
//   spring-boot-starter-thymeleaf  -> html/template (stdlib) if server-side
//                                     HTML rendering is still required; the
//                                     migrated handlers currently return JSON.
//   spring-boot-devtools           -> no equivalent; live-reload is handled by
//                                     external tools like github.com/air-verse/air
//                                     during development only.
//   lombok (compile-time)          -> no equivalent needed; Go structs are
//                                     written explicitly. Getters/setters,
//                                     constructors, and builders that Lombok
//                                     generated are replaced by exported fields
//                                     and NewXxx constructor functions.
//   spring-boot-starter-test       -> the stdlib `testing` package plus
//                                     github.com/stretchr/testify for
//                                     assertions/mocks (see the migrated
//                                     *_test files in internal/smartcontact).
//
// ---------------------------------------------------------------------------
// Build / packaging
// ---------------------------------------------------------------------------
//   spring-boot-maven-plugin (repackage into an executable fat JAR)
//     => `go build -o bin/smartcontact ./cmd/server`
//        produces a single self-contained native binary; no JVM or fat JAR is
//        involved. The application entry point is the composition root
//        cmd/server/main.go, which wires Config -> *sql.DB -> services ->
//        UserController -> Router (all already migrated).
//
// ---------------------------------------------------------------------------
// Spring Boot version note (from the migration debate)
// ---------------------------------------------------------------------------
// The parent POM pins spring-boot-starter-parent 2.7.14, which is built on
// Spring Framework 5.x. Post-Spring-5, StringHttpMessageConverter defaults to
// UTF-8 (pre-Spring-5 defaulted to ISO-8859-1). The Go handlers should
// therefore set text/plain responses with charset=utf-8 to preserve the
// original wire behaviour. JSON responses are UTF-8 by definition, so the
// existing JSON handlers already match.
//
// REQUIRES MANUAL REVIEW: create go.mod using the mapping above and run
// `go mod tidy`. This file is a placeholder for build metadata only.
package internal
