```go
// Package internal_test documents and validates the migration mapping described
// in internal/pom.xml.go.  Because the migrated file contains no executable
// code (it is documentation-only), the tests validate the migration contract
// expressed in the package-level doc comment: dependency mapping, toolchain
// equivalences, and the invariants listed in the behavioral specs.
package internal_test

import (
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// pomFileDoc returns the package-level doc comment of internal/pom.xml.go so
// that every test can inspect the migration notes without duplicating the path.
func pomFileDoc(t *testing.T) string {
	t.Helper()

	// Locate the file relative to this test source file.
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")

	// This test lives at internal/pom.xml_test.go;
	// the target lives at internal/pom.xml.go in the same directory.
	dir := filepath.Dir(thisFile)
	targetPath := filepath.Join(dir, "pom.xml.go")

	src, err := os.ReadFile(targetPath)
	require.NoError(t, err, "pom.xml.go must be readable")
	return string(src)
}

// parsedPackageDoc returns the parsed package-level documentation string via
// go/parser so we can assert on individual sentences independently of
// formatting changes.
func parsedPackageDoc(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(thisFile)
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return fi.Name() == "pom.xml.go"
	}, parser.ParseComments)
	require.NoError(t, err)
	require.NotEmpty(t, pkgs, "parser must find at least one package")

	for _, pkg := range pkgs {
		d := doc.New(pkg, dir, 0)
		return d.Doc
	}
	return ""
}

// ---------------------------------------------------------------------------
// Table-driven tests
// ---------------------------------------------------------------------------

// TestMigrationNotePresent verifies the high-level migration preamble that
// corresponds to the "Maven Project Configuration" spec: the file must exist,
// be non-empty, and declare itself as a migration artifact.
func TestMigrationNotePresent(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "file declares itself as a Maven build-configuration artifact",
			contains: "pom.xml is a Maven build-configuration artifact",
		},
		{
			name:     "file states no runtime behavior exists",
			contains: "NO runtime behavior",
		},
		{
			name:     "file states it is safe to delete after review",
			contains: "safe to delete once the migration has been reviewed",
		},
		{
			name:     "MIGRATION_NOTE header is present",
			contains: "MIGRATION_NOTE",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestGoEquivalentDependencies validates the dependency mapping table that
// replaces the original Maven <dependencies> block.  Each row maps a Spring
// Boot starter to its Go ecosystem counterpart.
func TestGoEquivalentDependencies(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name            string
		mavenDep        string
		goEquivalent    string
		additionalCheck string
	}{
		{
			name:         "spring-boot-starter-web maps to net/http and chi router",
			mavenDep:     "spring-boot-starter-web",
			goEquivalent: "net/http",
		},
		{
			name:         "spring-boot-starter-web also maps to chi",
			mavenDep:     "spring-boot-starter-web",
			goEquivalent: "github.com/go-chi/chi/v5",
		},
		{
			name:         "spring-boot-starter-data-jpa maps to database/sql",
			mavenDep:     "spring-boot-starter-data-jpa",
			goEquivalent: "database/sql",
		},
		{
			name:         "spring-boot-starter-data-jpa also maps to sqlx",
			mavenDep:     "spring-boot-starter-data-jpa",
			goEquivalent: "github.com/jmoiron/sqlx",
		},
		{
			name:         "mysql-connector-j is dropped in favour of pgx",
			mavenDep:     "mysql-connector-j",
			goEquivalent: "github.com/jackc/pgx/v5",
		},
		{
			name:         "spring-boot-starter-validation maps to validator",
			mavenDep:     "spring-boot-starter-validation",
			goEquivalent: "github.com/go-playground/validator/v10",
		},
		{
			name:         "spring-boot-starter-thymeleaf maps to html/template",
			mavenDep:     "spring-boot-starter-thymeleaf",
			goEquivalent: "html/template",
		},
		{
			name:         "spring-boot-starter-test maps to testing stdlib",
			mavenDep:     "spring-boot-starter-test",
			goEquivalent: "testing",
		},
		{
			name:         "spring-boot-starter-test also maps to testify",
			mavenDep:     "spring-boot-starter-test",
			goEquivalent: "github.com/stretchr/testify",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.mavenDep,
				"Maven dependency name should appear in the mapping table")
			assert.Contains(t, content, tc.goEquivalent,
				"Go equivalent should appear in the mapping table")
		})
	}
}

// TestMysqlDroppedPostgresAdopted validates the invariant that mysql-connector-j
// is explicitly dropped and that the migration note warns against porting the
// MySQL driver.
func TestMysqlDroppedPostgresAdopted(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "DROPPED keyword appears next to mysql-connector-j",
			contains: "DROPPED",
		},
		{
			name:     "target DB is declared as PostgreSQL",
			contains: "PostgreSQL",
		},
		{
			name:     "file warns against porting MySQL driver",
			contains: "Do NOT port the MySQL driver",
		},
		{
			name:     "file mentions pgx stdlib driver",
			contains: "pgx stdlib driver",
		},
		{
			name:     "file mentions dollar-sign placeholders as Postgres dialect indicator",
			contains: "$1 placeholders",
		},
		{
			name:     "file mentions RETURNING id as Postgres dialect indicator",
			contains: "RETURNING id",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestLombokMigrationNotes validates the invariants around lombok:
// no annotation processing in Go, and the explicit instruction to write
// NewXxx constructor functions.
func TestLombokMigrationNotes(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "lombok entry appears in the dependency mapping",
			contains: "lombok",
		},
		{
			name:     "note states Go has no annotation processing",
			contains: "Go has no annotation",
		},
		{
			name:     "note instructs writing explicit structs and constructors",
			contains: "Write structs/constructors",
		},
		{
			name:     "note mentions NewXxx naming convention",
			contains: "NewXxx functions",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestDevtoolsAndAirEquivalent validates that spring-boot-devtools has no
// direct equivalent and that air / go run is suggested.
func TestDevtoolsAndAirEquivalent(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "devtools entry appears in the mapping",
			contains: "spring-boot-devtools",
		},
		{
			name:     "no equivalent noted for devtools",
			contains: "No equivalent",
		},
		{
			name:     "air is suggested as hot-reload tool",
			contains: "air",
		},
		{
			name:     "go run is mentioned as an alternative",
			contains: "go run",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestSpringBootMavenPluginEquivalent validates the build plugin migration:
// spring-boot-maven-plugin -> go build producing a self-contained binary.
func TestSpringBootMavenPluginEquivalent(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "spring-boot-maven-plugin appears in mapping",
			contains: "spring-boot-maven-plugin",
		},
		{
			name:     "go build is the stated equivalent",
			contains: "go build",
		},
		{
			name:     "static binary is mentioned as the output",
			contains: "static binary",
		},
		{
			name:     "no fat-JAR packaging is needed",
			contains: "no fat-JAR packaging needed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestJavaVersionMappedToGoDirective validates the Java 17 -> go.mod directive
// mapping.
func TestJavaVersionMappedToGoDirective(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "Java version 17 is referenced",
			contains: "Java version: 17",
		},
		{
			name:     "go directive in go.mod is the equivalent",
			contains: "go.mod",
		},
		{
			name:     "example go directive version is present",
			contains: "go 1.22",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestConfigMigrationReference validates that the file references the already-
// migrated config package and the environment variables it reads.
func TestConfigMigrationReference(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "config.go is mentioned as already migrated",
			contains: "internal/config/config.go",
		},
		{
			name:     "DATABASE_URL env var is listed",
			contains: "DATABASE_URL",
		},
		{
			name:     "PORT env var is listed",
			contains: "PORT",
		},
		{
			name:     "JWT_SECRET env var is listed",
			contains: "JWT_SECRET",
		},
		{
			name:     "viper is mentioned as the env reader",
			contains: "viper",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, content, tc.contains)
		})
	}
}

// TestPackageDeclaration validates that the file belongs to the correct package.
func TestPackageDeclaration(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(thisFile)
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return fi.Name() == "pom.xml.go"
	}, parser.ParseComments)
	require.NoError(t, err)

	tests := []struct {
		name     string
		wantPkg  string
	}{
		{
			name:    "pom.xml.go belongs to package internal",
			wantPkg: "internal",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, found := pkgs[tc.wantPkg]
			assert.True(t, found, "package %q should be present in parsed output", tc.wantPkg)
		})
	}
}

// TestFileContainsNoBuildConstraints ensures the documentation file does not
// accidentally carry build constraints that would exclude it from normal builds.
func TestFileContainsNoBuildConstraints(t *testing.T) {
	content := pomFileDoc(t)

	tests := []struct {
		name        string
		absent      string
		description string
	}{
		{
			name:        "no go:build constraint present",
			absent:      "//go:build",
			description: "pom.xml.go is a documentation file and must not be build-constrained",
		},
		{
			name:        "no legacy +build tag present",
			absent:      "// +build",
			description: "legacy build tags must not appear in the documentation file",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.NotContains(t, content, tc.absent, tc.description)
		})
	}
}

// TestInvariantsFromBehavioralSpecs validates every invariant listed in the
// behavioral spec that can be asserted from the migration-note content.
func TestInvariantsFromBehavioralSpecs(t *testing.T) {
	content := pomFileDoc(t)

	// These invariants are checked as textual presence because the file is
	// purely documentation — it preserves the original POM metadata as prose.
	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		// modelVersion invariant — referenced indirectly through the Maven 4.0.0 mention.
		{
			name:     "Maven 4.0.0 project descriptor is referenced",
			contains: "Maven",
			reason:   "invariant: POM defines a valid Maven 4.0.0 project descriptor",
		},
		// parent invariant
		{