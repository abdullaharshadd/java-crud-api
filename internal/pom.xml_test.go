```go
package internal_test

import (
	"go/ast"
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

// sourceContent holds the raw source of pom.xml.go for inspection tests.
func sourceContent(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")
	dir := filepath.Dir(thisFile)
	src, err := os.ReadFile(filepath.Join(dir, "pom.xml.go"))
	require.NoError(t, err, "pom.xml.go must be readable")
	return string(src)
}

// parsedFile returns the AST of pom.xml.go.
func parsedFile(t *testing.T) *ast.File {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(thisFile)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(dir, "pom.xml.go"), nil, parser.ParseComments)
	require.NoError(t, err, "pom.xml.go must parse without error")
	return f
}

// ---------------------------------------------------------------------------
// Package-level invariants
// ---------------------------------------------------------------------------

func TestPackageDeclaration(t *testing.T) {
	t.Parallel()
	f := parsedFile(t)
	assert.Equal(t, "internal", f.Name.Name,
		"file must belong to package internal")
}

func TestNoRuntimeSymbolsDeclared(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		description string
		check       func(t *testing.T, f *ast.File)
	}{
		{
			name:        "no function declarations",
			description: "pom.xml.go is declarative; it must export no functions",
			check: func(t *testing.T, f *ast.File) {
				for _, decl := range f.Decls {
					_, isFn := decl.(*ast.FuncDecl)
					assert.False(t, isFn,
						"pom.xml.go must not declare any functions – found one")
				}
			},
		},
		{
			name:        "no variable declarations",
			description: "pom.xml.go is declarative; it must export no vars",
			check: func(t *testing.T, f *ast.File) {
				for _, decl := range f.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok {
						continue
					}
					assert.NotEqual(t, token.VAR, gd.Tok,
						"pom.xml.go must not declare any variables")
				}
			},
		},
		{
			name:        "no type declarations",
			description: "pom.xml.go is declarative; it must export no types",
			check: func(t *testing.T, f *ast.File) {
				for _, decl := range f.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok {
						continue
					}
					assert.NotEqual(t, token.TYPE, gd.Tok,
						"pom.xml.go must not declare any types")
				}
			},
		},
		{
			name:        "no const declarations",
			description: "pom.xml.go is declarative; it must export no consts",
			check: func(t *testing.T, f *ast.File) {
				for _, decl := range f.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok {
						continue
					}
					assert.NotEqual(t, token.CONST, gd.Tok,
						"pom.xml.go must not declare any constants")
				}
			},
		},
		{
			name:        "no import declarations",
			description: "pom.xml.go must not import any packages at compile time",
			check: func(t *testing.T, f *ast.File) {
				assert.Empty(t, f.Imports,
					"pom.xml.go must have no import statements")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := parsedFile(t)
			tc.check(t, f)
		})
	}
}

// ---------------------------------------------------------------------------
// Maven artifact coordinate invariants (documented in the file comments)
// ---------------------------------------------------------------------------

func TestMavenArtifactCoordinates(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name    string
		needle  string
		message string
	}{
		{
			name:    "groupId documented",
			needle:  "com.smartContact",
			message: "file must document Maven groupId com.smartContact",
		},
		{
			name:    "artifactId documented",
			needle:  "SmartContact",
			message: "file must document Maven artifactId SmartContact",
		},
		{
			name:    "version documented",
			needle:  "0.0.1-SNAPSHOT",
			message: "file must document Maven version 0.0.1-SNAPSHOT",
		},
		{
			name:    "modelVersion documented",
			needle:  "4.0.0",
			message: "file must document Maven modelVersion 4.0.0",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.message)
		})
	}
}

// ---------------------------------------------------------------------------
// Spring Boot parent / Java version invariants
// ---------------------------------------------------------------------------

func TestSpringBootParentVersionInvariant(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		{
			name:   "spring boot parent version 2.7.14",
			needle: "2.7.14",
			msg:    "file must document spring-boot-starter-parent version 2.7.14",
		},
		{
			name:   "java version 17 documented",
			needle: "17",
			msg:    "file must document Java 17 target",
		},
		{
			name:   "spring-boot-maven-plugin documented",
			needle: "spring-boot-maven-plugin",
			msg:    "file must document the spring-boot-maven-plugin",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// Dependency mapping documentation
// ---------------------------------------------------------------------------

func TestDependencyMappings(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		// Web / MVC
		{
			name:   "spring-boot-starter-web documented",
			needle: "spring-boot-starter-web",
			msg:    "file must document spring-boot-starter-web dependency",
		},
		{
			name:   "thymeleaf documented",
			needle: "thymeleaf",
			msg:    "file must document Thymeleaf dependency",
		},
		{
			name:   "chi router Go mapping",
			needle: "go-chi/chi",
			msg:    "file must document go-chi/chi as the Go routing equivalent",
		},
		{
			name:   "html/template Go mapping",
			needle: "html/template",
			msg:    "file must document html/template as the server-side rendering equivalent",
		},
		// JPA / persistence
		{
			name:   "spring-boot-starter-data-jpa documented",
			needle: "spring-boot-starter-data-jpa",
			msg:    "file must document spring-boot-starter-data-jpa dependency",
		},
		{
			name:   "database/sql Go mapping",
			needle: "database/sql",
			msg:    "file must document database/sql as the Go persistence equivalent",
		},
		{
			name:   "sqlx Go mapping",
			needle: "sqlx",
			msg:    "file must document sqlx as a Go persistence helper",
		},
		// Validation
		{
			name:   "spring-boot-starter-validation documented",
			needle: "spring-boot-starter-validation",
			msg:    "file must document spring-boot-starter-validation dependency",
		},
		// MySQL / PostgreSQL
		{
			name:   "mysql-connector documented",
			needle: "mysql-connector",
			msg:    "file must document the mysql-connector-j dependency",
		},
		{
			name:   "PostgreSQL migration note",
			needle: "PostgreSQL",
			msg:    "file must document the migration to PostgreSQL",
		},
		{
			name:   "pgx Go driver mapping",
			needle: "pgx",
			msg:    "file must document pgx as the Go PostgreSQL driver",
		},
		// Lombok
		{
			name:   "lombok documented",
			needle: "lombok",
			msg:    "file must document lombok dependency",
		},
		// Devtools
		{
			name:   "spring-boot-devtools documented",
			needle: "spring-boot-devtools",
			msg:    "file must document spring-boot-devtools dependency",
		},
		{
			name:   "air hot-reload Go mapping",
			needle: "air",
			msg:    "file must document air as an optional dev hot-reload tool",
		},
		// Test
		{
			name:   "spring-boot-starter-test documented",
			needle: "spring-boot-starter-test",
			msg:    "file must document spring-boot-starter-test dependency",
		},
		{
			name:   "testing stdlib Go mapping",
			needle: "testing",
			msg:    "file must document the stdlib testing package as equivalent",
		},
		{
			name:   "testify Go mapping",
			needle: "testify",
			msg:    "file must document testify as a Go test assertion library",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// Lombok exclusion invariant
// ---------------------------------------------------------------------------

func TestLombokExcludedFromFinalArtifact(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		{
			name:   "lombok exclusion or compile-time note documented",
			needle: "compile",
			msg:    "file must document that Lombok is a compile-time-only dependency",
		},
		{
			name:   "lombok no runtime note",
			needle: "lombok",
			msg:    "file must mention lombok in the context of no Go equivalent",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// MySQL vs PostgreSQL migration note
// ---------------------------------------------------------------------------

func TestMySQLToPostgreSQLMigrationNote(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		check  func(src string) bool
		msg    string
	}{
		{
			name: "MySQL is NOT recommended",
			check: func(src string) bool {
				// The file must warn not to use the MySQL driver.
				return strings.Contains(src, "NOT MySQL") ||
					strings.Contains(src, "not MySQL") ||
					strings.Contains(src, "Do not port")
			},
			msg: "file must explicitly warn against using the MySQL driver in Go",
		},
		{
			name: "PostgreSQL placeholder style documented",
			check: func(src string) bool {
				return strings.Contains(src, "$1") ||
					strings.Contains(src, "RETURNING")
			},
			msg: "file must document PostgreSQL placeholder style ($1, RETURNING)",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tc.check(src), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// go.mod authoritative source note
// ---------------------------------------------------------------------------

func TestGoModAuthoritativeNote(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		{
			name:   "go.mod mentioned as authoritative",
			needle: "go.mod",
			msg:    "file must reference go.mod as the authoritative dependency source",
		},
		{
			name:   "go directive note",
			needle: "go directive",
			msg:    "file must document how to set the Go toolchain version via the go directive",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// Migration notes presence
// ---------------------------------------------------------------------------

func TestMigrationNotesPresent(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		{
			name:   "MIGRATION_NOTE marker present",
			needle: "MIGRATION_NOTE",
			msg:    "file must contain at least one MIGRATION_NOTE comment",
		},
		{
			name:   "source file reference",
			needle: "pom.xml",
			msg:    "file must reference the original pom.xml source",
		},
		{
			name:   "no runtime symbols note",
			needle: "no runtime",
			msg:    "file must state it declares no runtime symbols",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(src, tc.needle), tc.msg)
		})
	}
}

// ---------------------------------------------------------------------------
// Build scenario: packaging produces a single binary (go build analogue)
// ---------------------------------------------------------------------------

func TestGoBuildEquivalentDocumented(t *testing.T) {
	t.Parallel()
	src := sourceContent(t)

	tests := []struct {
		name   string
		needle string
		msg    string
	}{
		{
			name:   "go build documented",
			needle: "go build",
			msg:    "file must document `go build` as the packaging equivalent",
		},
		{
			name:   "statically-linked binary note",
			needle: "statically-linked",
			msg:    "file must document that go build produces a statically-linked binary",
		},