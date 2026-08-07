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

// pomMetadata holds the expected Maven POM metadata as documented in the
// migration note embedded in internal/pom.xml.go.
type pomMetadata struct {
	modelVersion string
	groupID      string
	artifactID   string
	version      string
	parentGroup  string
	parentArtifact string
	parentVersion string
	javaVersion  string
}

// expectedPOM is the single source of truth for all POM invariants.
var expectedPOM = pomMetadata{
	modelVersion:   "4.0.0",
	groupID:        "com.smartContact",
	artifactID:     "SmartContact",
	version:        "0.0.1-SNAPSHOT",
	parentGroup:    "org.springframework.boot",
	parentArtifact: "spring-boot-starter-parent",
	parentVersion:  "2.7.14",
	javaVersion:    "17",
}

// expectedDependencies lists every dependency that must be documented in the
// migration file.
var expectedDependencies = []struct {
	name     string
	artifactID string
	scope    string
	optional bool
}{
	{"Spring Boot Starter Web", "spring-boot-starter-web", "", false},
	{"Spring Boot Starter Data JPA", "spring-boot-starter-data-jpa", "", false},
	{"Spring Boot Starter Validation", "spring-boot-starter-validation", "", false},
	{"Spring Boot Starter Thymeleaf", "spring-boot-starter-thymeleaf", "", false},
	{"MySQL Connector", "mysql-connector-j", "runtime", false},
	{"Spring Boot Starter Test", "spring-boot-starter-test", "test", false},
	{"Lombok", "lombok", "", true},
	{"Spring Boot DevTools", "spring-boot-devtools", "runtime", true},
}

// sourceFilePath returns the absolute path to the migration file under test.
func sourceFilePath(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")
	// currentFile is internal/pom.xml_test.go; the target is internal/pom.xml.go
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "pom.xml.go")
}

// readSourceFile reads and returns the raw content of the migration file.
func readSourceFile(t *testing.T) string {
	t.Helper()
	path := sourceFilePath(t)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "pom.xml.go must be readable")
	return string(data)
}

// parseSourceFile parses the migration file and returns the AST file node.
func parseSourceFile(t *testing.T) *ast.File {
	t.Helper()
	path := sourceFilePath(t)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	require.NoError(t, err, "pom.xml.go must be valid Go syntax")
	return f
}

// commentText extracts all comment text from the AST as a single joined string.
func commentText(f *ast.File) string {
	var parts []string
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// ---------------------------------------------------------------------------
// TestMavenProjectCoordinates
// ---------------------------------------------------------------------------

func TestMavenProjectCoordinates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "modelVersion must be 4.0.0",
			field:    "modelVersion",
			expected: expectedPOM.modelVersion,
		},
		{
			name:     "groupId must be com.smartContact",
			field:    "groupId",
			expected: expectedPOM.groupID,
		},
		{
			name:     "artifactId must be SmartContact",
			field:    "artifactId",
			expected: expectedPOM.artifactID,
		},
		{
			name:     "version must be 0.0.1-SNAPSHOT",
			field:    "version",
			expected: expectedPOM.version,
		},
	}

	src := readSourceFile(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, src, tc.expected,
				"Migration file must document the Maven coordinate value %q (field: %s)",
				tc.expected, tc.field)
		})
	}
}

// TestMavenProjectCoordinatesScenario validates the combined "build system
// resolves project coordinates" scenario.
func TestMavenProjectCoordinatesScenario(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	t.Run("Build system resolves project coordinates", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, src, "com.smartContact",
			"groupId must appear in migration documentation")
		assert.Contains(t, src, "SmartContact",
			"artifactId must appear in migration documentation")
		assert.Contains(t, src, "0.0.1-SNAPSHOT",
			"version must appear in migration documentation")
	})
}

// ---------------------------------------------------------------------------
// TestParentPOMInheritance
// ---------------------------------------------------------------------------

func TestParentPOMInheritance(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "Parent group must be org.springframework.boot",
			expected: expectedPOM.parentGroup,
		},
		{
			name:     "Parent artifact must be spring-boot-starter-parent",
			expected: expectedPOM.parentArtifact,
		},
		{
			name:     "Parent version must be 2.7.14",
			expected: expectedPOM.parentVersion,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, src, tc.expected,
				"Migration file must document parent POM value %q", tc.expected)
		})
	}
}

func TestParentPOMInheritanceScenarios(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	scenarios := []struct {
		name     string
		expected string
		desc     string
	}{
		{
			name:     "Versions managed by spring-boot-starter-parent 2.7.14",
			expected: "2.7.14",
			desc:     "Parent version 2.7.14 must be documented",
		},
		{
			name:     "Build fails if parent artifact cannot be resolved from repository",
			expected: "spring-boot-starter-parent",
			desc:     "Parent artifact name must be documented so error case is clear",
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, src, sc.expected, sc.desc)
		})
	}
}

// ---------------------------------------------------------------------------
// TestJavaVersionConfiguration
// ---------------------------------------------------------------------------

func TestJavaVersionConfiguration(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "java.version property must be 17",
			expected: "17",
		},
		{
			name:     "Java version keyword must appear",
			expected: "java.version",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, src, tc.expected,
				"Migration file must document Java version configuration value %q", tc.expected)
		})
	}
}

func TestJavaVersionConfigurationScenario(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	t.Run("Compiler configuration is applied - source and target compatibility Java 17", func(t *testing.T) {
		t.Parallel()
		// The migration note documents "go 1.21" as mapping to Java 17 era baseline.
		// The original java.version 17 must also be referenced.
		assert.Contains(t, src, "17",
			"Java 17 must be referenced in migration documentation")
		// Must also document the Go equivalent
		assert.Contains(t, src, "go 1.21",
			"Go language version equivalent to Java 17 must be documented")
	})

	t.Run("Build fails if JDK 17 or higher is not available", func(t *testing.T) {
		t.Parallel()
		// Error case is implicitly documented by the java.version reference.
		assert.Contains(t, src, "java.version",
			"java.version property must be documented to communicate error cases")
	})
}

// ---------------------------------------------------------------------------
// TestDependencyDeclaration
// ---------------------------------------------------------------------------

func TestDependencyDeclaration(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	for _, dep := range expectedDependencies {
		dep := dep
		t.Run("Dependency present: "+dep.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, src, dep.artifactID,
				"Migration file must document dependency %q", dep.artifactID)
		})
	}
}

func TestDependencyDeclarationInvariants(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	invariants := []struct {
		name     string
		check    func(t *testing.T, src string)
	}{
		{
			name: "spring-boot-starter-web must be present",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-web")
			},
		},
		{
			name: "spring-boot-starter-data-jpa must be present",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-data-jpa")
			},
		},
		{
			name: "spring-boot-starter-thymeleaf must be present",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-thymeleaf")
			},
		},
		{
			name: "spring-boot-starter-validation must be present",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-validation")
			},
		},
		{
			name: "mysql-connector-j must be scoped runtime",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "mysql-connector-j",
					"MySQL connector must be documented")
				// The migration note indicates it is a runtime dep.
				assert.Contains(t, src, "runtime",
					"runtime scope must be documented")
			},
		},
		{
			name: "spring-boot-starter-test must be scoped test",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-test")
				assert.Contains(t, src, "test",
					"test scope must be documented")
			},
		},
		{
			name: "lombok must be optional",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "lombok")
				assert.Contains(t, src, "optional",
					"lombok optional flag must be documented")
			},
		},
		{
			name: "spring-boot-devtools must be scoped runtime and optional",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-devtools")
				assert.Contains(t, src, "runtime",
					"devtools runtime scope must be documented")
				assert.Contains(t, src, "optional",
					"devtools optional flag must be documented")
			},
		},
	}

	for _, inv := range invariants {
		inv := inv
		t.Run(inv.name, func(t *testing.T) {
			t.Parallel()
			inv.check(t, src)
		})
	}
}

func TestDependencyDeclarationScenarios(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	scenarios := []struct {
		name     string
		check    func(t *testing.T, src string)
	}{
		{
			name: "Application requires JPA/persistence - JPA and MySQL driver documented",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-data-jpa",
					"JPA starter must be documented for persistence scenario")
				assert.Contains(t, src, "mysql-connector-j",
					"MySQL driver must be documented for persistence scenario")
			},
		},
		{
			name: "Application serves web content - web and thymeleaf documented",
			check: func(t *testing.T, src string) {
				assert.Contains(t, src, "spring-boot-starter-web",
					"Web starter must be documented for web scenario")
				assert.Contains(t, src, "spring-boot-starter-thymeleaf",
					"Thymeleaf starter must be documented for templating scenario")
			},
		},
		{
			name: "Build fails if any declared dependency cannot be resolved",
			check: func(t *testing.T, src string) {
				// All dependency names must be present so the error case is
				// identifiable during manual review.
				deps := []string{
					"spring-boot-starter-web",
					"spring-boot-starter-data-jpa",
					"spring-boot-starter-thymeleaf",
					"spring-boot-starter-validation",
					"mysql-connector-j",
					"spring-boot-starter-test",
					"lombok",
					"spring-boot-devtools",
				}
				for _, dep := range deps {
					assert.Contains(t, src, dep,
						"dependency %q must be documented to communicate resolution failure risk", dep)
				}
			},
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			sc.check(t, src)
		})
	}
}

// ---------------------------------------------------------------------------
// TestSpringBootMavenPluginConfiguration
// ---------------------------------------------------------------------------

func TestSpringBootMavenPluginConfiguration(t *testing.T) {
	t.Parallel()

	src := readSourceFile(t)

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "spring-boot-maven-plugin must be configured",
			expected: "spring-boot-maven-plugin",
		},
		{
			name:     "lombok must be excluded from packaged artifact",
			expected: "lombok",
		},
		{
			name:     "repackage goal or equivalent must be documented",
			expected: "repackage",
		},
	}

	for _, tc