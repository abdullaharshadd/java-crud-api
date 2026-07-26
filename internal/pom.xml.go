// Package internal documents the Maven Project Object Model (POM) that
// described the original Java/Spring Boot build for the SmartContact
// application. This file has NO runtime behavior in the Go port; it exists
// purely to preserve dependency-inventory and build-parity metadata so that
// anyone reproducing or auditing the original build knows exactly which
// frameworks and versions were declared.
//
// MIGRATION_NOTE: The source artifact was `pom.xml`, a Maven build
// configuration. Go projects do not use Maven; the Go toolchain and modules
// (go.mod / go.sum) fully replace this mechanism. There is no idiomatic Go
// equivalent for a POM, so the original coordinates and the recommended Go
// replacements are captured here as exported constants and a documented
// mapping table for audit purposes only. This file can be safely deleted
// once build-parity notes have been recorded elsewhere (e.g. in the repo
// README or go.mod).
package internal

// Maven project coordinates from the original pom.xml. Preserved verbatim for
// audit and traceability.
const (
	// ProjectGroupID is the Maven groupId of the original artifact.
	ProjectGroupID = "com.smartContact"
	// ProjectArtifactID is the Maven artifactId of the original artifact.
	ProjectArtifactID = "SmartContact"
	// ProjectVersion is the Maven version of the original artifact.
	ProjectVersion = "0.0.1-SNAPSHOT"
	// ProjectName is the human-readable project name.
	ProjectName = "SmartContact"
	// ProjectDescription is the human-readable project description.
	ProjectDescription = "smart Contact project "
	// SpringBootParentVersion is the spring-boot-starter-parent version that
	// managed transitive dependency versions via its BOM.
	SpringBootParentVersion = "2.7.14"
	// JavaVersion is the JDK version the original project targeted.
	JavaVersion = "17"
)

// DependencyMapping records how a single original Maven dependency maps onto
// its idiomatic Go replacement (or lack thereof). It is documentation only.
type DependencyMapping struct {
	// MavenGroupID is the original Maven groupId.
	MavenGroupID string
	// MavenArtifactID is the original Maven artifactId.
	MavenArtifactID string
	// Scope is the original Maven scope (empty means "compile").
	Scope string
	// GoReplacement is the recommended Go module path, or a note explaining
	// why no direct equivalent is needed.
	GoReplacement string
}

// DependencyInventory captures the full dependency list from the original
// pom.xml alongside the Go modules that replace each one. Per the migration
// debate, this POM is reference-only for dependency inventory:
//
//   - Spring Web MVC        -> github.com/go-chi/chi/v5 (HTTP routing)
//   - Spring Data JPA       -> github.com/jmoiron/sqlx  (thin SQL mapping)
//   - Bean Validation       -> github.com/go-playground/validator/v10
//   - MySQL connector       -> github.com/go-sql-driver/mysql
//
// Thymeleaf server-side rendering maps onto Go's standard html/template
// package; Lombok, devtools, and the Spring Boot test starter have no Go
// equivalent because their responsibilities are handled by the Go language,
// toolchain, and standard testing package respectively.
var DependencyInventory = []DependencyMapping{
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-devtools",
		Scope:           "runtime",
		GoReplacement:   "No equivalent needed; use `go run`/`air` for live reload during development.",
	},
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-starter-data-jpa",
		Scope:           "",
		GoReplacement:   "github.com/jmoiron/sqlx (with database/sql) replaces JPA/Hibernate ORM.",
	},
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-starter-validation",
		Scope:           "",
		GoReplacement:   "github.com/go-playground/validator/v10 replaces Jakarta Bean Validation.",
	},
	{
		MavenGroupID:    "org.projectlombok",
		MavenArtifactID: "lombok",
		Scope:           "",
		GoReplacement:   "No equivalent needed; Go structs require no boilerplate generation.",
	},
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-starter-thymeleaf",
		Scope:           "",
		GoReplacement:   "Standard library html/template replaces Thymeleaf server-side rendering.",
	},
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-starter-web",
		Scope:           "",
		GoReplacement:   "github.com/go-chi/chi/v5 with net/http replaces embedded Tomcat + Spring MVC.",
	},
	{
		MavenGroupID:    "com.mysql",
		MavenArtifactID: "mysql-connector-j",
		Scope:           "runtime",
		GoReplacement:   "github.com/go-sql-driver/mysql (registered as a database/sql driver).",
	},
	{
		MavenGroupID:    "org.springframework.boot",
		MavenArtifactID: "spring-boot-starter-test",
		Scope:           "test",
		GoReplacement:   "Standard library testing package plus github.com/stretchr/testify.",
	},
}
