```go
package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProjectCoordinates validates that the Maven project coordinates are
// correctly defined as exported constants.
func TestProjectCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "groupId matches com.smartContact",
			got:      ProjectGroupID,
			expected: "com.smartContact",
		},
		{
			name:     "artifactId matches SmartContact",
			got:      ProjectArtifactID,
			expected: "SmartContact",
		},
		{
			name:     "version matches 0.0.1-SNAPSHOT",
			got:      ProjectVersion,
			expected: "0.0.1-SNAPSHOT",
		},
		{
			name:     "project name matches SmartContact",
			got:      ProjectName,
			expected: "SmartContact",
		},
		{
			name:     "project description is set",
			got:      ProjectDescription,
			expected: "smart Contact project ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

// TestSpringBootParentVersion validates that the Spring Boot parent version is
// correctly pinned to 2.7.14.
func TestSpringBootParentVersion(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "parent version is fixed at 2.7.14",
			got:      SpringBootParentVersion,
			expected: "2.7.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

// TestJavaVersionConfiguration validates that the Java version is correctly
// set to 17.
func TestJavaVersionConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "java.version property is 17",
			got:      JavaVersion,
			expected: "17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

// TestDependencyInventoryLength verifies that the dependency inventory
// captures all eight declared Maven dependencies.
func TestDependencyInventoryLength(t *testing.T) {
	assert.Equal(t, 8, len(DependencyInventory),
		"DependencyInventory should contain exactly 8 entries matching pom.xml")
}

// TestDependencyInventoryContainsRequiredDependencies validates all declared
// Maven dependencies and their Go replacement notes are present and correct.
func TestDependencyInventoryContainsRequiredDependencies(t *testing.T) {
	type lookupKey struct {
		groupID    string
		artifactID string
	}

	// Build a lookup map for convenient assertion.
	depMap := make(map[lookupKey]DependencyMapping, len(DependencyInventory))
	for _, d := range DependencyInventory {
		depMap[lookupKey{d.MavenGroupID, d.MavenArtifactID}] = d
	}

	tests := []struct {
		name              string
		groupID           string
		artifactID        string
		expectedScope     string
		replacementSubstr string
	}{
		{
			name:              "spring-boot-starter-data-jpa is present",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-starter-data-jpa",
			expectedScope:     "",
			replacementSubstr: "github.com/jmoiron/sqlx",
		},
		{
			name:              "spring-boot-starter-validation is present",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-starter-validation",
			expectedScope:     "",
			replacementSubstr: "github.com/go-playground/validator/v10",
		},
		{
			name:              "spring-boot-starter-thymeleaf is present",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-starter-thymeleaf",
			expectedScope:     "",
			replacementSubstr: "html/template",
		},
		{
			name:              "spring-boot-starter-web is present",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-starter-web",
			expectedScope:     "",
			replacementSubstr: "github.com/go-chi/chi/v5",
		},
		{
			name:              "spring-boot-devtools is runtime-scoped",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-devtools",
			expectedScope:     "runtime",
			replacementSubstr: "air",
		},
		{
			name:              "lombok is present as compile-time-only",
			groupID:           "org.projectlombok",
			artifactID:        "lombok",
			expectedScope:     "",
			replacementSubstr: "No equivalent needed",
		},
		{
			name:              "mysql-connector-j is runtime-scoped",
			groupID:           "com.mysql",
			artifactID:        "mysql-connector-j",
			expectedScope:     "runtime",
			replacementSubstr: "github.com/go-sql-driver/mysql",
		},
		{
			name:              "spring-boot-starter-test is test-scoped",
			groupID:           "org.springframework.boot",
			artifactID:        "spring-boot-starter-test",
			expectedScope:     "test",
			replacementSubstr: "testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := lookupKey{tt.groupID, tt.artifactID}
			dep, exists := depMap[key]
			assert.True(t, exists, "dependency %s:%s should be in DependencyInventory", tt.groupID, tt.artifactID)
			if !exists {
				return
			}
			assert.Equal(t, tt.expectedScope, dep.Scope,
				"dependency %s:%s should have scope %q", tt.groupID, tt.artifactID, tt.expectedScope)
			assert.Contains(t, dep.GoReplacement, tt.replacementSubstr,
				"GoReplacement for %s:%s should mention %q", tt.groupID, tt.artifactID, tt.replacementSubstr)
		})
	}
}

// TestDependencyMappingFields validates that every DependencyMapping in the
// inventory has non-empty MavenGroupID, MavenArtifactID, and GoReplacement.
func TestDependencyMappingFields(t *testing.T) {
	for i, dep := range DependencyInventory {
		t.Run(dep.MavenArtifactID, func(t *testing.T) {
			assert.NotEmpty(t, dep.MavenGroupID,
				"entry[%d] MavenGroupID must not be empty", i)
			assert.NotEmpty(t, dep.MavenArtifactID,
				"entry[%d] MavenArtifactID must not be empty", i)
			assert.NotEmpty(t, dep.GoReplacement,
				"entry[%d] GoReplacement must not be empty", i)
		})
	}
}

// TestScopedDependencies verifies the scope assignments for each dependency
// against the invariants stated in the behavioral specs.
func TestScopedDependencies(t *testing.T) {
	tests := []struct {
		name          string
		groupID       string
		artifactID    string
		expectedScope string
		description   string
	}{
		{
			name:          "devtools runtime scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-devtools",
			expectedScope: "runtime",
			description:   "devtools must be runtime-scoped (optional in Maven sense)",
		},
		{
			name:          "mysql connector runtime scope",
			groupID:       "com.mysql",
			artifactID:    "mysql-connector-j",
			expectedScope: "runtime",
			description:   "mysql-connector-j must be runtime-scoped",
		},
		{
			name:          "test starter test scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-starter-test",
			expectedScope: "test",
			description:   "spring-boot-starter-test must be test-scoped only",
		},
		{
			name:          "jpa compile scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-starter-data-jpa",
			expectedScope: "",
			description:   "JPA starter has default (compile) scope",
		},
		{
			name:          "validation compile scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-starter-validation",
			expectedScope: "",
			description:   "validation starter has default (compile) scope",
		},
		{
			name:          "thymeleaf compile scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-starter-thymeleaf",
			expectedScope: "",
			description:   "thymeleaf starter has default (compile) scope",
		},
		{
			name:          "web compile scope",
			groupID:       "org.springframework.boot",
			artifactID:    "spring-boot-starter-web",
			expectedScope: "",
			description:   "web starter has default (compile) scope",
		},
		{
			name:          "lombok compile scope",
			groupID:       "org.projectlombok",
			artifactID:    "lombok",
			expectedScope: "",
			description:   "lombok has default (compile) scope and is optional",
		},
	}

	// Build lookup once.
	type key struct{ g, a string }
	depMap := make(map[key]DependencyMapping, len(DependencyInventory))
	for _, d := range DependencyInventory {
		depMap[key{d.MavenGroupID, d.MavenArtifactID}] = d
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, ok := depMap[key{tt.groupID, tt.artifactID}]
			assert.True(t, ok, "%s: dependency must exist in inventory", tt.description)
			if ok {
				assert.Equal(t, tt.expectedScope, dep.Scope,
					"%s: scope mismatch for %s:%s", tt.description, tt.groupID, tt.artifactID)
			}
		})
	}
}

// TestNoTestDependencyLeakIntoCompile validates the global invariant that
// test-scoped dependencies are not present with compile or runtime scope.
func TestNoTestDependencyLeakIntoCompile(t *testing.T) {
	tests := []struct {
		name       string
		groupID    string
		artifactID string
	}{
		{
			name:       "spring-boot-starter-test must not be compile-scoped",
			groupID:    "org.springframework.boot",
			artifactID: "spring-boot-starter-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dep := range DependencyInventory {
				if dep.MavenGroupID == tt.groupID && dep.MavenArtifactID == tt.artifactID {
					assert.Equal(t, "test", dep.Scope,
						"%s:%s must be test-scoped and must not leak into compile or runtime",
						tt.groupID, tt.artifactID)
				}
			}
		})
	}
}

// TestLombokIsCompileTimeOnly validates that Lombok is not listed as a runtime
// or test dependency — ensuring it is a compile-time-only aid per the global
// invariants.
func TestLombokIsCompileTimeOnly(t *testing.T) {
	t.Run("lombok must not be runtime or test scoped", func(t *testing.T) {
		for _, dep := range DependencyInventory {
			if dep.MavenGroupID == "org.projectlombok" && dep.MavenArtifactID == "lombok" {
				assert.NotEqual(t, "runtime", dep.Scope,
					"lombok must not be runtime-scoped; it is compile-time only")
				assert.NotEqual(t, "test", dep.Scope,
					"lombok must not be test-scoped; it is compile-time only")
				assert.Contains(t, dep.GoReplacement, "No equivalent needed",
					"lombok GoReplacement note should indicate no bundling needed in Go")
			}
		}
	})
}

// TestSpringBootVersion_GlobalInvariant checks the global invariant that the
// project is a Spring Boot 2.7.14 application.
func TestSpringBootVersion_GlobalInvariant(t *testing.T) {
	tests := []struct {
		name          string
		constant      string
		expectedValue string
	}{
		{
			name:          "Spring Boot parent version is 2.7.14",
			constant:      SpringBootParentVersion,
			expectedValue: "2.7.14",
		},
		{
			name:          "Java target version is 17",
			constant:      JavaVersion,
			expectedValue: "17",
		},
		{
			name:          "Project artifact is SmartContact",
			constant:      ProjectArtifactID,
			expectedValue: "SmartContact",
		},
		{
			name:          "Project version is 0.0.1-SNAPSHOT",
			constant:      ProjectVersion,
			expectedValue: "0.0.1-SNAPSHOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedValue, tt.constant)
		})
	}
}

// TestDependencyInventorySpringBootStartersHaveConsistentGroupID validates
// that all Spring Boot starters share the same groupId as the parent, ensuring
// version management consistency.
func TestDependencyInventorySpringBootStartersHaveConsistentGroupID(t *testing.T) {
	springBootStarters := []string{
		"spring-boot-starter-data-jpa",
		"spring-boot-starter-validation",
		"spring-boot-starter-thymeleaf",
		"spring-boot-starter-web",
		"spring-boot-starter-test",
		"spring-boot-devtools",
	}

	type key struct{ g, a string }
	depMap := make(map[key]DependencyMapping, len(DependencyInventory))
	for _, d := range DependencyInventory {
		depMap[key{d.MavenGroupID, d.MavenArtifactID}] = d
	}

	for _, artifactID := range springBootStarters {
		artifactID := artifactID // capture
		t.Run("groupId for "+artifactID, func(t *testing.T) {
			dep, ok := depMap[key{"org.springframework.boot", artifactID}]
			assert.True(t, ok, "%s must be present in DependencyInventory", artifactID)
			if ok {
				assert.Equal(t, "org.springframework.boot", dep.MavenGroupID,
					"%s must have groupId org.springframework.boot