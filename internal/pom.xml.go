// Package internal documents the build and dependency configuration that was
// previously expressed as a Maven POM (pom.xml) for the SmartContact Spring Boot
// application.
//
// MIGRATION_NOTE: pom.xml is not source code — it is Maven's declarative build
// descriptor. It has NO runtime business logic to translate; it describes:
//   - the build tool and parent POM (spring-boot-starter-parent 2.7.14),
//   - the language level (Java 17),
//   - third-party dependencies (JPA, validation, web, Thymeleaf, MySQL, Lombok),
//   - and packaging plugins (spring-boot-maven-plugin producing a fat JAR).
//
// Go has no equivalent single artifact. The idiomatic replacement is split
// across several well-defined mechanisms:
//
//   1. Dependency management  -> go.mod / go.sum (the Go module system).
//   2. Build & packaging      -> `go build` / `go install` producing a single
//                                statically-linked binary (the fat-JAR analogue
//                                is native to Go and needs no plugin).
//   3. Language version        -> the `go` directive in go.mod.
//
// Because none of the Maven concepts map to a compilable Go declaration, this
// file intentionally contains ONLY documentation plus a machine-readable
// summary of the intended module dependencies, so a human can reproduce the
// build with the correct Go equivalents. It declares no runtime types or
// functions (there is nothing executable to declare) and must not be confused
// with the composition root, which lives in cmd/smartcontact.
//
// Below is the mapping from each Maven dependency to its idiomatic Go
// counterpart. This is a guide for maintaining go.mod, NOT auto-generated code.
package internal

// DependencyMapping describes how a single Maven dependency from the original
// pom.xml corresponds (or does not correspond) to a Go module. It exists purely
// for documentation/tooling and carries no runtime behaviour.
type DependencyMapping struct {
	// MavenArtifact is the original Maven groupId:artifactId.
	MavenArtifact string
	// GoModule is the recommended Go module import path, or empty if the
	// capability is provided by the Go standard library or has no equivalent.
	GoModule string
	// Note explains the migration decision for this dependency.
	Note string
}

// MavenToGoDependencies enumerates the original pom.xml dependencies and their
// idiomatic Go replacements. Keep this in sync with go.mod when dependencies
// change. It is exported so build tooling or documentation generators can
// consume it, but it is intentionally inert at runtime.
//
// MIGRATION_NOTE: The target database is PostgreSQL (the original used MySQL).
// The MySQL driver dependency is therefore deliberately NOT carried over —
// use github.com/jackc/pgx or the standard database/sql + github.com/lib/pq
// driver instead, with $1,$2 positional placeholders and RETURNING id on INSERT.
var MavenToGoDependencies = []DependencyMapping{
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-parent:2.7.14",
		GoModule:      "",
		Note:          "Parent POM version management has no Go analogue; go.mod pins exact versions directly.",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-devtools",
		GoModule:      "",
		Note:          "Hot-reload dev tooling; use `air` or `go run` locally. Not a runtime dependency.",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-data-jpa",
		GoModule:      "database/sql (stdlib)",
		Note:          "No ORM/JPA. Repositories use raw database/sql with explicit PostgreSQL queries (see internal/smartcontact/repository).",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-validation",
		GoModule:      "github.com/go-playground/validator/v10",
		Note:          "Replaces Jakarta Bean Validation annotations with explicit struct-tag validation, or hand-written validation in the handler layer.",
	},
	{
		MavenArtifact: "org.projectlombok:lombok",
		GoModule:      "",
		Note:          "Compile-time boilerplate generation (getters/setters/builders) is unnecessary in Go; fields are used directly.",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-thymeleaf",
		GoModule:      "html/template (stdlib)",
		Note:          "Server-side rendering; use html/template for HTML views if the MVC UI is retained.",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-web",
		GoModule:      "net/http (stdlib) + github.com/go-chi/chi/v5",
		Note:          "Embedded server + routing; net/http provides the server, chi provides idiomatic routing/middleware (see RegisterRoutes in internal/smartcontact/handler).",
	},
	{
		MavenArtifact: "com.mysql:mysql-connector-j",
		GoModule:      "github.com/jackc/pgx/v5 (or github.com/lib/pq)",
		Note:          "MIGRATION_NOTE: Target DB is PostgreSQL, not MySQL. Do NOT use a MySQL driver; wire the Postgres driver into OpenDB (internal/resources/application.properties.go).",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-starter-test",
		GoModule:      "testing (stdlib) + github.com/stretchr/testify",
		Note:          "Table-driven tests via the stdlib testing package; testify for assertions/mocks.",
	},
	{
		MavenArtifact: "org.springframework.boot:spring-boot-maven-plugin",
		GoModule:      "",
		Note:          "Fat-JAR packaging is native to `go build` (single static binary); no plugin needed.",
	},
}

// GoLanguageVersion records the Java 17 language level from the original POM's
// <java.version> property, mapped to the recommended minimum Go toolchain. Set
// the actual minimum via the `go` directive in go.mod.
//
// MIGRATION_NOTE: Java 17 (LTS) does not map to a specific Go version; this is a
// documentation marker only. Use a currently-supported Go release (>= 1.21).
const GoLanguageVersion = "go1.21"

// ModulePath is the intended Go module path, replacing the Maven
// groupId:artifactId coordinate com.smartContact:SmartContact. It should match
// the module directive at the top of go.mod.
const ModulePath = "github.com/smartcontact/smartcontact"
