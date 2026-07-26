// Package internal documents build/dependency configuration that was originally
// expressed as a Maven POM file (pom.xml) for the SmartContact application.
//
// MIGRATION_NOTE: pom.xml is a Maven build-configuration file, not application
// logic. It has NO runtime behavior to migrate. A Go program built with the
// standard `go` toolchain (go build / go test / go modules) does not use Maven,
// so there is no direct functional equivalent. The Go equivalent of a POM is a
// go.mod file plus the Go build tooling; that lives at the module root, not in
// a generated .go source file.
//
// The original coordinates and dependency intent are preserved below as exported
// constants purely for documentation and traceability. They are NOT consumed by
// any Go build step.
//
// Manual review guidance for the porting effort:
//   - Spring Boot 2.7.14 pulls in Jackson for JSON. In Go, use encoding/json
//     (or a chosen library) explicitly. Confirm date/time serialization: the
//     Java side may emit epoch-millis for java.util.Date vs ISO-8601 for
//     java.time.LocalDateTime. Match the wire format deliberately in Go
//     (e.g. custom time.Time (un)marshalling) rather than assuming defaults.
//   - spring-boot-starter-data-jpa (Hibernate) => replace with database/sql +
//     a Postgres driver (e.g. github.com/jackc/pgx). NOTE the DATABASE MIGRATION:
//     the source targets MySQL (mysql-connector-j), but the target datastore is
//     PostgreSQL. Use $1,$2,... placeholders and `RETURNING id` for inserts;
//     do NOT port MySQL dialect or AUTO_INCREMENT semantics.
//   - spring-boot-starter-validation => use a Go validation library
//     (e.g. github.com/go-playground/validator) or explicit checks.
//   - spring-boot-starter-thymeleaf => use html/template for server-side views.
//   - spring-boot-starter-web => use net/http (optionally chi/gin) for routing.
//   - lombok => not applicable; Go has no annotation processing.
//   - spring-boot-devtools => not applicable; use `air`/`reflex` for live reload
//     during development if desired.
package internal

// Project metadata originally declared in pom.xml. These are documentation-only
// constants and are not referenced by the Go build.
const (
	// MavenGroupID is the original Maven groupId of the source project.
	MavenGroupID = "com.smartContact"

	// MavenArtifactID is the original Maven artifactId of the source project.
	MavenArtifactID = "SmartContact"

	// MavenVersion is the original project version from pom.xml.
	MavenVersion = "0.0.1-SNAPSHOT"

	// ProjectName is the human-readable project name from pom.xml.
	ProjectName = "SmartContact"

	// ProjectDescription is the project description from pom.xml.
	ProjectDescription = "smart Contact project"

	// SpringBootVersion is the Spring Boot parent version that governed the
	// original opinionated dependency versions (Jackson, Hibernate, etc.).
	// Retained for traceability when verifying serialization/ORM behavior.
	SpringBootVersion = "2.7.14"

	// JavaVersion is the source language level the original project compiled
	// against. Recorded for reference only.
	JavaVersion = "17"
)
