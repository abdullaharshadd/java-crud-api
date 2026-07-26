```go
package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Constants – project coordinate tests
// ---------------------------------------------------------------------------

func TestProjectCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "ProjectGroupID matches Maven groupId",
			got:      ProjectGroupID,
			expected: "com.smartContact",
		},
		{
			name:     "ProjectArtifactID matches Maven artifactId",
			got:      ProjectArtifactID,
			expected: "SmartContact",
		},
		{
			name:     "ProjectVersion matches Maven version",
			got:      ProjectVersion,
			expected: "0.0.1-SNAPSHOT",
		},
		{
			name:     "ProjectName matches Maven name",
			got:      ProjectName,
			expected: "SmartContact",
		},
		{
			name:     "ProjectDescription matches Maven description",
			got:      ProjectDescription,
			expected: "smart Contact project",
		},
		{
			name:     "JavaVersion records Java 17 target",
			got:      JavaVersion,
			expected: "17",
		},
		{
			name:     "SpringBootParentVersion records Spring Boot BOM 2.7.14",
			got:      SpringBootParentVersion,
			expected: "2.7.14",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.got)
		})
	}
}

// ---------------------------------------------------------------------------
// RecommendedDependencies – existence, completeness and content tests
// ---------------------------------------------------------------------------

func TestRecommendedDependencies_ReturnsNonNilSlice(t *testing.T) {
	deps := RecommendedDependencies()
	assert.NotNil(t, deps, "RecommendedDependencies() must not return nil")
}

func TestRecommendedDependencies_CountMatchesPomDependencies(t *testing.T) {
	// The original pom.xml declares exactly 8 dependencies (devtools, data-jpa,
	// validation, lombok, thymeleaf, web, mysql-connector-j, starter-test).
	deps := RecommendedDependencies()
	assert.Len(t, deps, 8, "expected one mapping entry per pom.xml dependency")
}

// TestRecommendedDependencies_NoEmptyFields ensures every entry is fully
// populated – a partially filled entry would be a documentation regression.
func TestRecommendedDependencies_NoEmptyFields(t *testing.T) {
	deps := RecommendedDependencies()
	for i, d := range deps {
		t.Run(d.MavenArtifact, func(t *testing.T) {
			assert.NotEmpty(t, d.MavenArtifact,
				"entry[%d].MavenArtifact must not be empty", i)
			assert.NotEmpty(t, d.GoEquivalent,
				"entry[%d].GoEquivalent must not be empty", i)
			assert.NotEmpty(t, d.Notes,
				"entry[%d].Notes must not be empty", i)
		})
	}
}

// TestRecommendedDependencies_EachPomDependencyIsMapped validates the global
// invariants that describe every required pom.xml dependency.
func TestRecommendedDependencies_EachPomDependencyIsMapped(t *testing.T) {
	tests := []struct {
		invariant     string
		mavenArtifact string
	}{
		{
			invariant:     "spring-boot-starter-data-jpa must be mapped (JPA invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-starter-data-jpa",
		},
		{
			invariant:     "spring-boot-starter-validation must be mapped (validation invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-starter-validation",
		},
		{
			invariant:     "spring-boot-starter-thymeleaf must be mapped (Thymeleaf invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-starter-thymeleaf",
		},
		{
			invariant:     "spring-boot-starter-web must be mapped (web invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-starter-web",
		},
		{
			invariant:     "lombok must be mapped (optional dependency invariant)",
			mavenArtifact: "org.projectlombok:lombok",
		},
		{
			invariant:     "spring-boot-devtools must be mapped (optional runtime/devtools invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-devtools",
		},
		{
			invariant:     "mysql-connector-j must be mapped (runtime DB-connectivity invariant)",
			mavenArtifact: "com.mysql:mysql-connector-j",
		},
		{
			invariant:     "spring-boot-starter-test must be mapped (test scope invariant)",
			mavenArtifact: "org.springframework.boot:spring-boot-starter-test",
		},
	}

	deps := RecommendedDependencies()

	// Build a lookup set for O(1) membership checks.
	mapped := make(map[string]DependencyMapping, len(deps))
	for _, d := range deps {
		mapped[d.MavenArtifact] = d
	}

	for _, tc := range tests {
		t.Run(tc.invariant, func(t *testing.T) {
			_, ok := mapped[tc.mavenArtifact]
			assert.True(t, ok,
				"dependency %q not found in RecommendedDependencies()", tc.mavenArtifact)
		})
	}
}

// ---------------------------------------------------------------------------
// Per-dependency content invariants
// ---------------------------------------------------------------------------

func TestRecommendedDependencies_DevtoolsMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-devtools"]
	assert.True(t, ok, "devtools entry must exist")
	if ok {
		assert.Contains(t, d.Notes, "Hot-reload",
			"devtools notes must mention hot-reload behaviour")
	}
}

func TestRecommendedDependencies_DataJpaMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-starter-data-jpa"]
	assert.True(t, ok, "data-jpa entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "database/sql",
			"Go equivalent must reference database/sql")
		assert.Contains(t, d.Notes, "JpaRepository",
			"notes must mention JpaRepository semantics")
	}
}

func TestRecommendedDependencies_ValidationMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-starter-validation"]
	assert.True(t, ok, "validation entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "validator",
			"Go equivalent must reference a validator library")
		assert.Contains(t, d.Notes, "validate",
			"notes must describe validation strategy")
	}
}

func TestRecommendedDependencies_LombokMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.projectlombok:lombok"]
	assert.True(t, ok, "lombok entry must exist")
	if ok {
		// Lombok has no Go equivalent – the mapping must make that clear.
		assert.Contains(t, d.GoEquivalent, "none",
			"Go equivalent for lombok should state 'none required'")
		assert.Contains(t, d.Notes, "getter",
			"notes must explain what Lombok generated (getters/setters)")
	}
}

func TestRecommendedDependencies_ThymeleafMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-starter-thymeleaf"]
	assert.True(t, ok, "thymeleaf entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "html/template",
			"Go equivalent must reference html/template")
	}
}

func TestRecommendedDependencies_WebMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-starter-web"]
	assert.True(t, ok, "web entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "net/http",
			"Go equivalent must reference net/http")
		assert.Contains(t, d.Notes, "ControllerAdvice",
			"notes must mention @ControllerAdvice migration")
	}
}

func TestRecommendedDependencies_MySQLConnectorMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["com.mysql:mysql-connector-j"]
	assert.True(t, ok, "mysql-connector-j entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "mysql",
			"Go equivalent must reference a MySQL driver")
		assert.Contains(t, d.Notes, "DSN",
			"notes must mention DSN-based connection registration")
	}
}

func TestRecommendedDependencies_StarterTestMapping(t *testing.T) {
	deps := RecommendedDependencies()
	mapped := indexByArtifact(deps)

	d, ok := mapped["org.springframework.boot:spring-boot-starter-test"]
	assert.True(t, ok, "starter-test entry must exist")
	if ok {
		assert.Contains(t, d.GoEquivalent, "testing",
			"Go equivalent must reference the testing standard library")
		assert.Contains(t, d.GoEquivalent, "testify",
			"Go equivalent must mention testify")
		assert.Contains(t, d.Notes, "table-driven",
			"notes must mention table-driven tests per Go convention")
	}
}

// ---------------------------------------------------------------------------
// DependencyMapping struct field tests
// ---------------------------------------------------------------------------

func TestDependencyMapping_StructFields(t *testing.T) {
	tests := []struct {
		name          string
		mapping       DependencyMapping
		wantArtifact  string
		wantEquiv     string
		wantNotesFrag string
	}{
		{
			name: "fully populated mapping is preserved verbatim",
			mapping: DependencyMapping{
				MavenArtifact: "example:artifact",
				GoEquivalent:  "example/go/pkg",
				Notes:         "some migration note",
			},
			wantArtifact:  "example:artifact",
			wantEquiv:     "example/go/pkg",
			wantNotesFrag: "some migration note",
		},
		{
			name: "zero-value mapping has all empty fields",
			mapping: DependencyMapping{},
			wantArtifact:  "",
			wantEquiv:     "",
			wantNotesFrag: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantArtifact, tc.mapping.MavenArtifact)
			assert.Equal(t, tc.wantEquiv, tc.mapping.GoEquivalent)
			assert.Contains(t, tc.mapping.Notes, tc.wantNotesFrag)
		})
	}
}

// ---------------------------------------------------------------------------
// Idempotency – calling RecommendedDependencies() multiple times must return
// the same result (no hidden mutable state).
// ---------------------------------------------------------------------------

func TestRecommendedDependencies_Idempotent(t *testing.T) {
	first := RecommendedDependencies()
	second := RecommendedDependencies()

	assert.Equal(t, len(first), len(second),
		"repeated calls must return slices of equal length")

	for i := range first {
		assert.Equal(t, first[i], second[i],
			"entry[%d] must be identical across calls", i)
	}
}

// ---------------------------------------------------------------------------
// Global invariant: Spring Boot parent version 2.7.14
// ---------------------------------------------------------------------------

func TestSpringBootParentVersion_MatchesInvariant(t *testing.T) {
	assert.Equal(t, "2.7.14", SpringBootParentVersion,
		"global invariant: build must inherit from spring-boot-starter-parent 2.7.14")
}

// ---------------------------------------------------------------------------
// Global invariant: Java 17 target
// ---------------------------------------------------------------------------

func TestJavaVersion_MatchesInvariant(t *testing.T) {
	assert.Equal(t, "17", JavaVersion,
		"global invariant: build must target Java version 17")
}

// ---------------------------------------------------------------------------
// Global invariant: model version 4.0.0 is implicit in pom.xml; the artifact
// coordinates uniquely identify the project.
// ---------------------------------------------------------------------------

func TestProjectIdentity_GlobalInvariant(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"groupId", ProjectGroupID, "com.smartContact"},
		{"artifactId", ProjectArtifactID, "SmartContact"},
		{"version", ProjectVersion, "0.0.1-SNAPSHOT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.got,
				"global invariant: project identity %s must match pom.xml", tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// indexByArtifact converts a slice of DependencyMapping into a map keyed by
// MavenArtifact for O(1) lookups in test assertions.
func indexByArtifact(deps []DependencyMapping) map[string]DependencyMapping {
	m := make(map[string]DependencyMapping, len(deps))
	for _, d := range deps {
		m[d.MavenArtifact] = d
	}
	return m
}
```