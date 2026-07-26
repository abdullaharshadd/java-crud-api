// Package internal documents the migration of the Maven build manifest (pom.xml)
// to the Go ecosystem.
//
// MIGRATION_NOTE: pom.xml is a Maven Project Object Model — a *build manifest*,
// not application code. It has NO direct, executable Go equivalent. Go uses
// go.mod / go.sum plus the `go` toolchain for dependency management and
// reproducible builds. This file exists only to satisfy the migration mapping
// and to serve as documentation of the original build configuration and the
// framework behaviors that MUST be faithfully reproduced by hand in the Go
// port.
//
// REQUIRES MANUAL REVIEW:
//   - Create a proper go.mod (module path, Go version) instead of this file.
//   - Wire the equivalent Go libraries listed in the dependency mapping below.
//   - Delete this file once go.mod is authoritative.
package internal

// ---------------------------------------------------------------------------
// Project coordinates (informational only)
// ---------------------------------------------------------------------------

const (
	// ProjectGroupID mirrors the Maven <groupId>.
	ProjectGroupID = "com.smartContact"
	// ProjectArtifactID mirrors the Maven <artifactId>.
	ProjectArtifactID = "SmartContact"
	// ProjectVersion mirrors the Maven <version>.
	ProjectVersion = "0.0.1-SNAPSHOT"
	// ProjectName mirrors the Maven <name>.
	ProjectName = "SmartContact"
	// ProjectDescription mirrors the Maven <description>.
	ProjectDescription = "smart Contact project"
	// JavaVersion records the original JDK target. In Go the equivalent is the
	// `go` directive in go.mod (e.g. `go 1.22`).
	JavaVersion = "17"
	// SpringBootParentVersion records the Spring Boot BOM version whose
	// transitive versions were inherited by the original build.
	SpringBootParentVersion = "2.7.14"
)

// DependencyMapping documents how each Maven dependency maps onto the Go
// ecosystem. It is intended as a checklist for the human performing the port.
//
// MIGRATION_NOTE: These are recommendations, not automatic wiring. Add the
// chosen modules to go.mod with `go get`.
type DependencyMapping struct {
	// MavenArtifact is the original Maven groupId:artifactId.
	MavenArtifact string
	// GoEquivalent is the recommended idiomatic Go replacement.
	GoEquivalent string
	// Notes explains behavior that must be preserved by hand.
	Notes string
}

// RecommendedDependencies returns the Go equivalents for every dependency
// declared in the original pom.xml, along with the framework behaviors that
// must be replicated explicitly (Spring/JPA magic has no Go analogue).
func RecommendedDependencies() []DependencyMapping {
	return []DependencyMapping{
		{
			MavenArtifact: "org.springframework.boot:spring-boot-devtools",
			GoEquivalent:  "air (github.com/cosmtrek/air) or `go run` with a file watcher",
			Notes:         "Hot-reload dev tooling only; not a runtime dependency in Go.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-data-jpa",
			GoEquivalent:  "database/sql + github.com/jmoiron/sqlx (or gorm.io/gorm)",
			Notes:         "Replicate JpaRepository semantics manually: e.g. deleteById -> repo.DeleteByID(ctx, id) executing a DELETE and returning (rowsAffected, error). No implicit transactions.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-validation",
			GoEquivalent:  "github.com/go-playground/validator/v10",
			Notes:         "Replace javax.validation annotations with `validate:` struct tags; validate explicitly in handlers.",
		},
		{
			MavenArtifact: "org.projectlombok:lombok",
			GoEquivalent:  "none required",
			Notes:         "Lombok generated getters/setters/constructors. Go uses exported fields and NewXxx constructors; no code generation needed.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-thymeleaf",
			GoEquivalent:  "html/template (standard library)",
			Notes:         "Server-side HTML rendering. Port Thymeleaf templates to Go html/template.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-web",
			GoEquivalent:  "net/http + github.com/go-chi/chi/v5",
			Notes:         "Replace @Controller/@RestController with explicit route registration. Replicate @ControllerAdvice + DefaultErrorAttributes with centralized error-handling middleware producing structured JSON errors.",
		},
		{
			MavenArtifact: "com.mysql:mysql-connector-j",
			GoEquivalent:  "github.com/go-sql-driver/mysql",
			Notes:         "MySQL driver; register with database/sql via a DSN.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-test",
			GoEquivalent:  "testing (standard library) + github.com/stretchr/testify",
			Notes:         "Use table-driven tests; httptest for HTTP handlers.",
		},
	}
}

// MIGRATION_NOTE: The <build><plugins> section (spring-boot-maven-plugin
// producing an executable/fat JAR with lombok excluded) has no Go analogue.
// In Go, `go build ./cmd/smartcontact` produces a single self-contained
// static binary — the fat-JAR problem does not exist.
//
// REQUIRES MANUAL REVIEW: no HTTP routes are defined in pom.xml, so there is
// nothing to register here. Route registration belongs in the migrated web
// controllers (spring-boot-starter-web). Ensure @ControllerAdvice and
// DefaultErrorAttributes behavior are ported to chi middleware there.
