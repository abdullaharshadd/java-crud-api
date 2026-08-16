// Package internal documents the migration of the Maven build configuration
// (pom.xml) to the Go module system.
//
// MIGRATION_NOTE: pom.xml is a Maven build/dependency-management descriptor,
// not application logic. Go has no direct analogue that lives inside a .go
// source file — the equivalent responsibilities are split across:
//
//   - go.mod / go.sum       -> module identity + dependency versions
//                              (replaces <groupId>/<artifactId>/<version> and
//                               the spring-boot-starter-parent BOM)
//   - `go build` / `go install` -> the build lifecycle
//                              (replaces spring-boot-maven-plugin fat-JAR
//                               packaging; `go build` already produces a single
//                               statically-linked executable, so no repackage
//                               step is required)
//
// This file therefore contains NO runnable code — it exists to record the
// dependency mapping decisions so a human can reconcile them against the
// committed go.mod. It compiles as an empty (declaration-free) member of the
// `internal` package.
//
// -----------------------------------------------------------------------------
// Dependency mapping (Maven artifact -> Go module)
// -----------------------------------------------------------------------------
//
//   spring-boot-starter-web         -> github.com/go-chi/chi/v5
//                                      (net/http router + middleware; see
//                                       internal/smartcontact/handler and
//                                       internal/smartcontact/smartcontactapplication.go)
//
//   spring-boot-starter-validation  -> github.com/go-playground/validator/v10
//                                      (bean-validation / @Valid replacement;
//                                       surfaced via apperror.ValidationError)
//
//   spring-boot-starter-data-jpa    -> database/sql (stdlib) + repository
//                                      pattern in
//                                      internal/smartcontact/repository/userdao.go.
//                                      Hibernate's implicit ORM is replaced by
//                                      explicit SQL. Schema creation that JPA's
//                                      ddl-auto handled is now performed
//                                      explicitly at boot (see the repository
//                                      layer), targeting PostgreSQL.
//
//   mysql-connector-j (runtime)     -> github.com/jackc/pgx/v5/stdlib
//                                      MIGRATION_NOTE: the source targeted
//                                      MySQL, but the agreed target dialect for
//                                      this migration is PostgreSQL. The MySQL
//                                      JDBC driver is intentionally NOT mirrored;
//                                      use a Postgres driver + $1 placeholders +
//                                      RETURNING id. REVIEW: confirm the running
//                                      database is Postgres, not MySQL.
//
//   spring-boot-starter-thymeleaf   -> html/template (stdlib) IF server-side
//                                      rendering is still required. No Thymeleaf
//                                      templates were part of this migration's
//                                      scope; the handlers migrated so far
//                                      return JSON. REVIEW: port any .html
//                                      Thymeleaf templates to html/template if
//                                      the UI is needed.
//
//   spring-boot-devtools (runtime)  -> no equivalent. Live-reload during dev is
//                                      an external tool concern (e.g. `air` or
//                                      `reflex`); not a compiled dependency.
//
//   lombok (compile-time)           -> no equivalent and none needed. Go has no
//                                      annotation processor; getters/setters/
//                                      builders that Lombok generated are simply
//                                      not required — struct fields are accessed
//                                      directly and constructors are explicit
//                                      NewXxx functions.
//
//   spring-boot-starter-test        -> testing (stdlib) + github.com/stretchr/testify.
//                                      See internal/smartcontact/service/user_service_impl_test.go
//                                      and internal/smartcontact/smartcontactapplicationtests.go.
//
// java.version=17                   -> the `go` directive in go.mod pins the
//                                      language/toolchain version instead.
//
// artifactId=SmartContact, version=0.0.1-SNAPSHOT
//                                   -> the module path + a VCS tag / build
//                                      ldflags carry identity and version.
//
// ACTION REQUIRED (human): ensure go.mod declares the modules listed above.
// This documentation file cannot create go.mod on its own.
package internal
