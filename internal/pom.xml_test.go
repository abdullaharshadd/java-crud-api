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

// pomDocumentation holds the parsed documentation from the migration file.
type pomDocumentation struct {
	packageName string
	docComment  string
	hasCode     bool
}

// parsePomFile parses the pom.xml.go file and extracts its documentation.
func parsePomFile(t *testing.T) pomDocumentation {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to determine current file path")

	dir := filepath.Dir(currentFile)
	targetFile := filepath.Join(dir, "pom.xml.go")

	src, err := os.ReadFile(targetFile)
	require.NoError(t, err, "pom.xml.go must be readable")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, targetFile, src, parser.ParseComments)
	require.NoError(t, err, "pom.xml.go must be valid Go")

	doc := ""
	if f.Doc != nil {
		doc = f.Doc.Text()
	}

	// Check for any declarations (functions, types, vars, consts)
	hasCode := len(f.Decls) > 0

	return pomDocumentation{
		packageName: f.Name.Name,
		docComment:  doc,
		hasCode:     hasCode,
	}
}

// TestPomXmlGoFileExists verifies the migration file is present and parseable.
func TestPomXmlGoFileExists(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(currentFile)
	targetFile := filepath.Join(dir, "pom.xml.go")

	_, err := os.Stat(targetFile)
	assert.NoError(t, err, "pom.xml.go must exist in the internal package")
}

// TestPomXmlGoPackageDeclaration verifies the package name.
func TestPomXmlGoPackageDeclaration(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name            string
		expectedPackage string
	}{
		{
			name:            "file belongs to internal package",
			expectedPackage: "internal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedPackage, doc.packageName,
				"pom.xml.go must be in package %q", tc.expectedPackage)
		})
	}
}

// TestPomXmlGoContainsNoRunableCode verifies the file is documentation-only.
func TestPomXmlGoContainsNoRunableCode(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name        string
		expectCode  bool
		description string
	}{
		{
			name:        "no declarations in migration documentation file",
			expectCode:  false,
			description: "pom.xml.go must contain no runnable code — it is a migration documentation file only",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectCode, doc.hasCode, tc.description)
		})
	}
}

// TestMavenBuildConfigurationDocumented verifies Maven build configuration
// metadata is recorded in the documentation comment.
func TestMavenBuildConfigurationDocumented(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name           string
		searchString   string
		description    string
		caseSensitive  bool
	}{
		// Invariant: modelVersion is 4.0.0 (indirectly referenced via "Maven")
		{
			name:          "maven migration note present",
			searchString:  "pom.xml",
			description:   "must document the Maven POM migration",
			caseSensitive: true,
		},
		// Invariant: parent is spring-boot-starter-parent version 2.7.14
		{
			name:          "spring-boot-starter-parent version documented",
			searchString:  "2.7.14",
			description:   "must document the Spring Boot 2.7.14 parent version",
			caseSensitive: true,
		},
		// Invariant: groupId is com.smartContact, artifactId is SmartContact, version is 0.0.1-SNAPSHOT
		{
			name:          "artifactId SmartContact documented",
			searchString:  "SmartContact",
			description:   "must document the SmartContact artifactId",
			caseSensitive: true,
		},
		{
			name:          "version 0.0.1-SNAPSHOT documented",
			searchString:  "0.0.1-SNAPSHOT",
			description:   "must document the 0.0.1-SNAPSHOT version",
			caseSensitive: true,
		},
		// Invariant: java.version property is 17
		{
			name:          "java version 17 documented",
			searchString:  "java.version=17",
			description:   "must document java.version=17",
			caseSensitive: true,
		},
		// spring-boot-starter-web mapping
		{
			name:          "spring-boot-starter-web mapping documented",
			searchString:  "spring-boot-starter-web",
			description:   "must document the spring-boot-starter-web dependency mapping",
			caseSensitive: true,
		},
		// spring-boot-starter-validation mapping
		{
			name:          "spring-boot-starter-validation mapping documented",
			searchString:  "spring-boot-starter-validation",
			description:   "must document the spring-boot-starter-validation dependency mapping",
			caseSensitive: true,
		},
		// spring-boot-starter-data-jpa mapping
		{
			name:          "spring-boot-starter-data-jpa mapping documented",
			searchString:  "spring-boot-starter-data-jpa",
			description:   "must document the spring-boot-starter-data-jpa dependency mapping",
			caseSensitive: true,
		},
		// mysql-connector-j mapping and runtime scope note
		{
			name:          "mysql-connector-j mapping documented",
			searchString:  "mysql-connector-j",
			description:   "must document the mysql-connector-j dependency mapping",
			caseSensitive: true,
		},
		// spring-boot-starter-thymeleaf mapping
		{
			name:          "spring-boot-starter-thymeleaf mapping documented",
			searchString:  "spring-boot-starter-thymeleaf",
			description:   "must document the spring-boot-starter-thymeleaf dependency mapping",
			caseSensitive: true,
		},
		// Invariant: spring-boot-devtools is runtime-scoped and optional
		{
			name:          "spring-boot-devtools documented",
			searchString:  "spring-boot-devtools",
			description:   "must document the spring-boot-devtools dependency",
			caseSensitive: true,
		},
		// Invariant: lombok is optional and excluded from the repackaged artifact
		{
			name:          "lombok documented",
			searchString:  "lombok",
			description:   "must document the lombok dependency",
			caseSensitive: false,
		},
		// Invariant: spring-boot-starter-test is test-scoped
		{
			name:          "spring-boot-starter-test mapping documented",
			searchString:  "spring-boot-starter-test",
			description:   "must document the spring-boot-starter-test dependency mapping",
			caseSensitive: true,
		},
		// Go module equivalents documented
		{
			name:          "go-chi router replacement documented",
			searchString:  "github.com/go-chi/chi",
			description:   "must document the go-chi/chi replacement for spring-boot-starter-web",
			caseSensitive: true,
		},
		{
			name:          "go-playground/validator replacement documented",
			searchString:  "github.com/go-playground/validator",
			description:   "must document the go-playground/validator replacement for spring-boot-starter-validation",
			caseSensitive: true,
		},
		{
			name:          "pgx postgres driver replacement documented",
			searchString:  "pgx",
			description:   "must document the pgx driver replacing mysql-connector-j",
			caseSensitive: true,
		},
		{
			name:          "testify replacement for spring-boot-starter-test documented",
			searchString:  "testify",
			description:   "must document testify as replacement for spring-boot-starter-test",
			caseSensitive: true,
		},
		// Go module system equivalents for Maven concepts
		{
			name:          "go.mod mentioned as module management replacement",
			searchString:  "go.mod",
			description:   "must document go.mod as the replacement for Maven dependency management",
			caseSensitive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.caseSensitive {
				assert.Contains(t, doc.docComment, tc.searchString, tc.description)
			} else {
				assert.Contains(t, strings.ToLower(doc.docComment), strings.ToLower(tc.searchString), tc.description)
			}
		})
	}
}

// TestMavenBuildScenarios verifies documented build behavior expectations.
func TestMavenBuildScenarios(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name         string
		searchString string
		description  string
	}{
		{
			name:         "spring-boot-maven-plugin packaging documented",
			searchString: "spring-boot-maven-plugin",
			description:  "must document the spring-boot-maven-plugin packaging concern and its Go equivalent",
		},
		{
			name:         "fat-jar or executable artifact documented",
			searchString: "statically-linked",
			description:  "must document that go build produces a statically-linked executable replacing the fat JAR",
		},
		{
			name:         "lombok exclusion from repackaged artifact documented",
			searchString: "Lombok",
			description:  "must document Lombok exclusion behaviour",
		},
		{
			name:         "build lifecycle replacement documented",
			searchString: "go build",
			description:  "must document the go build command as lifecycle replacement",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, doc.docComment, tc.searchString, tc.description)
		})
	}
}

// TestDependencyScopeDocumentation verifies scope information is captured.
func TestDependencyScopeDocumentation(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name         string
		searchString string
		description  string
	}{
		{
			name:         "runtime scope noted for devtools",
			searchString: "spring-boot-devtools",
			description:  "spring-boot-devtools runtime scope must be noted",
		},
		{
			name:         "mysql runtime scope noted",
			searchString: "runtime",
			description:  "runtime scope for mysql-connector-j must be documented",
		},
		{
			name:         "test scope for spring-boot-starter-test noted",
			searchString: "testing",
			description:  "test scope for spring-boot-starter-test equivalent must be documented",
		},
		{
			name:         "optional dependency flag noted",
			searchString: "optional",
			description:  "optional flag for devtools/lombok must be documented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, doc.docComment, tc.searchString, tc.description)
		})
	}
}

// TestMigrationInvariantsDocumented validates the global invariants are captured.
func TestMigrationInvariantsDocumented(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name         string
		searchString string
		description  string
	}{
		{
			name:         "data-jpa starter invariant documented",
			searchString: "spring-boot-starter-data-jpa",
			description:  "required starter data-jpa must be documented as compile dependency",
		},
		{
			name:         "validation starter invariant documented",
			searchString: "spring-boot-starter-validation",
			description:  "required starter validation must be documented as compile dependency",
		},
		{
			name:         "thymeleaf starter invariant documented",
			searchString: "spring-boot-starter-thymeleaf",
			description:  "required starter thymeleaf must be documented",
		},
		{
			name:         "web starter invariant documented",
			searchString: "spring-boot-starter-web",
			description:  "required starter web must be documented as compile dependency",
		},
		{
			name:         "spring boot parent version invariant documented",
			searchString: "2.7.14",
			description:  "global invariant: Spring Boot 2.7.14 parent must be documented",
		},
		{
			name:         "java 17 toolchain invariant documented",
			searchString: "17",
			description:  "global invariant: Java 17 target must be documented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, doc.docComment, tc.searchString, tc.description)
		})
	}
}

// TestReviewAndActionNotes verifies the migration file documents human action items.
func TestReviewAndActionNotes(t *testing.T) {
	doc := parsePomFile(t)

	tests := []struct {
		name         string
		searchString string
		description  string
	}{
		{
			name:         "action required note for go.mod present",
			searchString: "ACTION REQUIRED",
			description:  "must include an ACTION REQUIRED note for humans to update go.mod",
		},
		{
			name:         "review note for mysql/postgres migration present",
			searchString: "REVIEW",
			description:  "must include REVIEW notes for MySQL to PostgreSQL migration decisions",
		},
		{
			name:         "migration note present",
			searchString: "MIGRATION_NOTE",
			description:  "must include MIGRATION_NOTE annotations explaining mapping decisions",
		},
		{
			name:         "postgres dialect migration noted",
			searchString: "PostgreSQL",
			description:  "must note the agreed target dialect is PostgreSQL, not MySQL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, doc.docComment, tc.searchString, tc.description)
		})
	}
}

// TestPomXmlGoFileIsValidGo verifies the file compiles as valid Go.
func TestPomXmlGoFileIsValidGo(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(currentFile)
	targetFile := filepath.Join(dir, "pom.xml.go")

	src, err := os.ReadFile(targetFile)
	require.NoError(t, err)

	tests := []struct {
		name     string
		src      []byte