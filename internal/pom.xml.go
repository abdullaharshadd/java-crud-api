// Package internal documents the build and dependency configuration for the
// SmartContact application.
//
// MIGRATION_NOTE: The source file (pom.xml) is a Maven Project Object Model.
// It is a build manifest, not executable code, and therefore has no direct
// runtime equivalent in Go. There is nothing to "run" here. This file exists
// only to document how the Maven build configuration maps onto the Go module
// ecosystem, so a human reviewer can reproduce the dependency and build setup.
//
// MIGRATION_NOTE: In Go, dependency management is handled by go.mod / go.sum
// (the analogue of <dependencies>) and the build is driven by the `go build`
// / `go test` toolchain (the analogue of the spring-boot-maven-plugin). The
// mappings below are advisory; the authoritative source of truth for the Go
// build is the module's go.mod file, NOT this documentation stub.
//
// Maven parent / packaging:
//   - spring-boot-starter-parent 2.7.14 (BOM / version management)
//       -> No Go equivalent. Go pins exact versions per-dependency in go.mod;
//          there is no inherited "parent POM" concept.
//   - spring-boot-maven-plugin (executable fat JAR)
//       -> `go build` already produces a single statically-linked binary,
//          so no repackaging plugin is required.
//   - java.version 17
//       -> Set the Go toolchain version via the `go` directive in go.mod.
//
// Dependency mapping (Java artifact -> recommended Go module):
//   - spring-boot-starter-web / thymeleaf (HTTP + server-rendered MVC)
//       -> github.com/go-chi/chi/v5 for routing (see handler/usercontroller.go
//          and smartcontactapplication.go, which already use chi). Server-side
//          HTML rendering, if still needed, maps to the stdlib html/template.
//   - spring-boot-starter-data-jpa (Hibernate ORM)
//       -> database/sql + github.com/jmoiron/sqlx (see repository/userdao.go).
//          NOTE: no ORM is used; queries are hand-written SQL.
//   - spring-boot-starter-validation (Bean Validation / JSR-380)
//       -> Explicit validation in model.Validate (see model/user.go), or
//          github.com/go-playground/validator/v10 if struct-tag validation
//          is preferred.
//   - com.mysql:mysql-connector-j (runtime JDBC driver)
//       -> IMPORTANT: the target database is PostgreSQL, NOT MySQL. Use
//          github.com/jackc/pgx/v5 (with the stdlib database/sql adapter
//          github.com/jackc/pgx/v5/stdlib) or github.com/lib/pq. Do not port
//          the MySQL driver. SQL uses $1, $2 placeholders and RETURNING id.
//   - org.projectlombok:lombok (compile-time boilerplate generation)
//       -> No Go equivalent needed. Go structs and explicit constructors
//          (e.g. model.NewUser) replace Lombok-generated getters/setters,
//          constructors, and builders.
//   - spring-boot-devtools (hot reload)
//       -> Optional dev-only tooling such as github.com/air-verse/air;
//          not a production dependency.
//   - spring-boot-starter-test (JUnit + Mockito + AssertJ)
//       -> The stdlib `testing` package with table-driven tests, plus
//          github.com/stretchr/testify for assertions/mocks if desired
//          (see service/service_test.go).
//
// MIGRATION_NOTE: This file intentionally declares no runtime symbols. All
// build/dependency intent is captured in go.mod and in the already-migrated
// source files referenced above. Human reviewers should verify that go.mod
// pins concrete versions equivalent to the Spring Boot 2.7.14 stack and that
// the PostgreSQL driver (not MySQL) is wired in.
package internal
