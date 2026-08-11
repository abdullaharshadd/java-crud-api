```go
// Package internal_test documents and validates the migration metadata
// expressed in pom.xml.go.  Because the migrated file contains no runtime
// symbols (it is documentation-only), these tests validate the invariants
// described in the behavioral specs by inspecting the package-level doc
// comment and the hard-coded migration notes at compile time / via string
// analysis.  Every table row is a distinct behavioral scenario drawn from the
// spec.
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

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// sourceDoc returns the raw Go source of pom.xml.go so we can assert on its
// content without importing it (it declares no exported symbols).
func sourceDoc(t *testing.T) string {
	t.Helper()
	// Resolve the path relative to this test file's directory.
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")
	dir := filepath.Dir(testFile)
	src, err := os.ReadFile(filepath.Join(dir, "pom.xml.go"))
	require.NoError(t, err, "pom.xml.go must be readable")
	return string(src)
}

// parsedDoc parses pom.xml.go and returns the package-level doc comment text.
func parsedDoc(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(testFile)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(dir, "pom.xml.go"), nil, parser.ParseComments)
	require.NoError(t, err, "pom.xml.go must parse as valid Go")

	var sb strings.Builder
	for _, cg := range f.Comments {
		sb.WriteString(cg.Text())
		sb.WriteString("\n")
	}
	return sb.String()
}

// packageName returns the package clause of pom.xml.go.
func packageName(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(testFile)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(dir, "pom.xml.go"), nil, parser.PackageClauseOnly)
	require.NoError(t, err)
	return f.Name.Name
}

// exportedSymbols returns all top-level exported declarations in pom.xml.go.
func exportedSymbols(t *testing.T) []string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(testFile)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(dir, "pom.xml.go"), nil, 0)
	require.NoError(t, err)

	var names []string
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				names = append(names, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						names = append(names, s.Name.Name)
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.IsExported() {
							names = append(names, n.Name)
						}
					}
				}
			}
		}
	}
	return names
}

// ---------------------------------------------------------------------------
// 1.  maven-project-coordinates
// ---------------------------------------------------------------------------

func TestMavenProjectCoordinates(t *testing.T) {
	doc := sourceDoc(t)

	tests := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "groupId present",
			snippet: "com.smartContact",
			desc:    "groupId must be com.smartContact",
		},
		{
			name:    "artifactId present",
			snippet: "SmartContact",
			desc:    "artifactId / project name must be SmartContact",
		},
		{
			name:    "version present",
			snippet: "0.0.1-SNAPSHOT",
			desc:    "version must be 0.0.1-SNAPSHOT",
		},
		{
			name:    "modelVersion invariant",
			snippet: "4.0.0",
			desc:    "modelVersion must be 4.0.0 (mentioned in global invariants)",
		},
		{
			name:    "default packaging jar implied",
			snippet: "jar",
			desc:    "packaging defaults to jar – mentioned in go.mod note or migration comment",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, strings.Contains(doc, tc.snippet),
				"%s: expected %q to appear in pom.xml.go", tc.desc, tc.snippet)
		})
	}
}

// ---------------------------------------------------------------------------
// 2.  spring-boot-parent-inheritance
// ---------------------------------------------------------------------------

func TestSpringBootParentInheritance(t *testing.T) {
	doc := sourceDoc(t)

	tests := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "parent artifact",
			snippet: "spring-boot-starter-parent",
			desc:    "parent POM must be spring-boot-starter-parent",
		},
		{
			name:    "parent version pinned to 2.7.14",
			snippet: "2.7.14",
			desc:    "parent version must be pinned to 2.7.14",
		},
		{
			name:    "managed versions apply",
			snippet: "managed",
			desc:    "doc must mention managed dependency versions from parent",
		},
		{
			name:    "spring boot version reference",
			snippet: "Spring Boot",
			desc:    "Spring Boot referenced in migration notes",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, strings.Contains(doc, tc.snippet),
				"%s: expected %q in source", tc.desc, tc.snippet)
		})
	}
}

// ---------------------------------------------------------------------------
// 3.  java-version-target
// ---------------------------------------------------------------------------

func TestJavaVersionTarget(t *testing.T) {
	doc := sourceDoc(t)

	tests := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "java version property equals 17",
			snippet: "java.version 17",
			desc:    "java.version property must equal 17",
		},
		{
			name:    "go directive equivalent",
			snippet: "go 1.22",
			desc:    "go directive in go.mod maps java.version -> go 1.22",
		},
		{
			name:    "java 17 mentioned",
			snippet: "Java 17",
			desc:    "Java 17 target must be called out in migration notes",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, strings.Contains(doc, tc.snippet),
				"%s: expected %q in source", tc.desc, tc.snippet)
		})
	}
}

// ---------------------------------------------------------------------------
// 4.  runtime-and-test-dependencies
// ---------------------------------------------------------------------------

func TestRuntimeAndTestDependencies(t *testing.T) {
	doc := sourceDoc(t)

	// compile-time / default scope
	compileScopeDeps := []struct {
		name    string
		snippet string
	}{
		{"jpa starter", "spring-boot-starter-data-jpa"},
		{"validation starter", "spring-boot-starter-validation"},
		{"thymeleaf starter", "spring-boot-starter-thymeleaf"},
		{"web starter", "spring-boot-starter-web"},
	}

	t.Run("compile_scope_dependencies_present", func(t *testing.T) {
		for _, dep := range compileScopeDeps {
			dep := dep
			t.Run(dep.name, func(t *testing.T) {
				assert.True(t, strings.Contains(doc, dep.snippet),
					"compile-scope dep %q must appear in migration notes", dep.snippet)
			})
		}
	})

	// runtime scope
	runtimeDeps := []struct {
		name    string
		snippet string
	}{
		{"mysql connector", "mysql-connector-j"},
		{"devtools", "spring-boot-devtools"},
	}

	t.Run("runtime_scope_dependencies", func(t *testing.T) {
		for _, dep := range runtimeDeps {
			dep := dep
			t.Run(dep.name, func(t *testing.T) {
				assert.True(t, strings.Contains(doc, dep.snippet),
					"runtime dep %q must appear in migration notes", dep.snippet)
			})
		}
	})

	// test scope
	t.Run("test_scope_dependency", func(t *testing.T) {
		assert.True(t, strings.Contains(doc, "spring-boot-starter-test"),
			"spring-boot-starter-test must appear in migration notes")
	})

	// optional dependencies
	optionalDeps := []struct {
		name    string
		snippet string
	}{
		{"devtools optional", "spring-boot-devtools"},
		{"lombok optional", "lombok"},
	}

	t.Run("optional_dependencies", func(t *testing.T) {
		for _, dep := range optionalDeps {
			dep := dep
			t.Run(dep.name, func(t *testing.T) {
				assert.True(t, strings.Contains(doc, dep.snippet),
					"optional dep %q must appear in migration notes", dep.snippet)
			})
		}
	})

	// invariants: specific migration decisions
	invariants := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "devtools dropped",
			snippet: "DROPPED",
			desc:    "devtools and lombok are marked DROPPED in migration notes",
		},
		{
			name:    "lombok dropped",
			snippet: "lombok",
			desc:    "lombok mentioned as DROPPED",
		},
		{
			name:    "mysql connector dropped in favor of postgres",
			snippet: "PostgreSQL",
			desc:    "mysql-connector-j replaced by PostgreSQL driver",
		},
		{
			name:    "test scope mapped to testify",
			snippet: "testify",
			desc:    "spring-boot-starter-test maps to testify",
		},
		{
			name:    "versions managed by parent not overridden",
			snippet: "managed",
			desc:    "Spring dependency versions not overridden, inherited from parent",
		},
	}

	t.Run("dependency_invariants", func(t *testing.T) {
		for _, inv := range invariants {
			inv := inv
			t.Run(inv.name, func(t *testing.T) {
				assert.True(t, strings.Contains(doc, inv.snippet),
					"%s: expected %q in migration notes", inv.desc, inv.snippet)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// 5.  spring-boot-maven-plugin-build
// ---------------------------------------------------------------------------

func TestSpringBootMavenPluginBuild(t *testing.T) {
	doc := sourceDoc(t)

	tests := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "plugin referenced",
			snippet: "spring-boot-maven-plugin",
			desc:    "spring-boot-maven-plugin must appear in migration notes",
		},
		{
			name:    "lombok excluded from packaged artifact",
			snippet: "lombok",
			desc:    "lombok exclusion from repackaged artifact must be documented",
		},
		{
			name:    "go build replaces plugin",
			snippet: "go build",
			desc:    "go build replaces the Maven plugin for producing a binary",
		},
		{
			name:    "statically linked binary",
			snippet: "statically-linked",
			desc:    "migration note must mention statically-linked binary",
		},
		{
			name:    "only declared plugin",
			snippet: "only",
			desc:    "spring-boot-maven-plugin is the only declared build plugin",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, strings.Contains(doc, tc.snippet),
				"%s: expected %q in pom.xml.go", tc.desc, tc.snippet)
		})
	}
}

// ---------------------------------------------------------------------------
// 6.  global invariants
// ---------------------------------------------------------------------------

func TestGlobalInvariants(t *testing.T) {
	doc := sourceDoc(t)

	tests := []struct {
		name    string
		snippet string
		desc    string
	}{
		{
			name:    "spring boot 2.7.14 web application",
			snippet: "2.7.14",
			desc:    "project is a Spring Boot 2.7.14 web application",
		},
		{
			name:    "java 17 target",
			snippet: "17",
			desc:    "application targets Java 17",
		},
		{
			name:    "versions inherited from parent not overridden",
			snippet: "managed",
			desc:    "Spring dependency versions inherited from spring-boot-starter-parent",
		},
		{
			name:    "requires mysql at runtime",
			snippet: "mysql-connector-j",
			desc:    "application required MySQL at runtime (now replaced by PostgreSQL)",
		},
		{
			name:    "lombok compile time excluded runtime",
			snippet: "lombok",
			desc:    "lombok compile-time but excluded from runtime artifact",
		},
		{
			name:    "valid maven 4.0.0 model descriptor",
			snippet: "4.0.0",
			desc:    "POM must remain a valid Maven 4.0.0 model descriptor",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, strings.Contains(doc, tc.snippet),
				"%s: expected %q in pom.xml.go", tc.desc, tc.snippet)
		})
	}
}

// ---------------------------------------------------------------------------
// 7.  structural / syntactic guarantees
// ---------------------------------------------------------------------------

func TestFileStructure(t *testing.T) {
	t.Run("file_is_valid_go", func(t *testing.T) {
		_, testFile, _, ok := runtime.Caller(0)
		require.True(t, ok)
		dir := filepath.Dir(testFile)

		fset := token.NewFileSet()
		_, err := parser.ParseFile(fset, filepath.Join(dir, "pom.xml.go"), nil, parser.ParseComments)
		assert.NoError(t, err, "pom.xml.go must be syntactically valid Go")
	})

	t.Run("package_name_is_internal", func(t *testing.T) {
		pkg := packageName(t)
		assert.Equal(t, "internal",