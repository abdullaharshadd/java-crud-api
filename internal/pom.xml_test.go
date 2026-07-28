```go
package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontact/smartcontact/internal"
)

// TestGoLanguageVersion validates the documented Go language version constant.
func TestGoLanguageVersion(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "GoLanguageVersion must map to a currently-supported Go release >= 1.21",
			expected: "go1.21",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, internal.GoLanguageVersion,
				"GoLanguageVersion must be >= go1.21 to satisfy the Java 17 LTS migration note")
		})
	}
}

// TestModulePath validates the Go module path constant.
func TestModulePath(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "ModulePath must match the intended Go module import path",
			expected: "github.com/smartcontact/smartcontact",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, internal.ModulePath,
				"ModulePath must replace the Maven groupId:artifactId coordinate com.smartContact:SmartContact")
		})
	}
}

// TestMavenToGoDependencies_NotEmpty ensures the mapping slice is populated.
func TestMavenToGoDependencies_NotEmpty(t *testing.T) {
	t.Run("MavenToGoDependencies must not be empty", func(t *testing.T) {
		assert.NotEmpty(t, internal.MavenToGoDependencies,
			"MavenToGoDependencies must enumerate at least one dependency mapping")
	})
}

// TestMavenToGoDependencies_AllFieldsPresent validates that every entry has a
// non-empty MavenArtifact and a non-empty Note.
func TestMavenToGoDependencies_AllFieldsPresent(t *testing.T) {
	for i, dep := range internal.MavenToGoDependencies {
		dep := dep // capture
		t.Run(dep.MavenArtifact, func(t *testing.T) {
			assert.NotEmpty(t, dep.MavenArtifact,
				"entry %d: MavenArtifact must not be empty", i)
			assert.NotEmpty(t, dep.Note,
				"entry %d: Note must not be empty for artifact %s", i, dep.MavenArtifact)
		})
	}
}

// expectedDependencies mirrors the pom.xml dependencies to ensure full coverage.
var expectedMavenArtifacts = []string{
	"org.springframework.boot:spring-boot-starter-parent:2.7.14",
	"org.springframework.boot:spring-boot-devtools",
	"org.springframework.boot:spring-boot-starter-data-jpa",
	"org.springframework.boot:spring-boot-starter-validation",
	"org.projectlombok:lombok",
	"org.springframework.boot:spring-boot-starter-thymeleaf",
	"org.springframework.boot:spring-boot-starter-web",
	"com.mysql:mysql-connector-j",
	"org.springframework.boot:spring-boot-starter-test",
	"org.springframework.boot:spring-boot-maven-plugin",
}

// TestMavenToGoDependencies_ContainsAllPomEntries ensures every pom.xml entry
// is represented in the mapping.
func TestMavenToGoDependencies_ContainsAllPomEntries(t *testing.T) {
	// build a set for O(1) lookup
	index := make(map[string]internal.DependencyMapping, len(internal.MavenToGoDependencies))
	for _, d := range internal.MavenToGoDependencies {
		index[d.MavenArtifact] = d
	}

	for _, artifact := range expectedMavenArtifacts {
		artifact := artifact
		t.Run("contains "+artifact, func(t *testing.T) {
			_, found := index[artifact]
			assert.True(t, found,
				"MavenToGoDependencies must contain an entry for Maven artifact %q", artifact)
		})
	}
}

// TestDependencyMapping_SpringBootParent validates the parent POM entry.
func TestDependencyMapping_SpringBootParent(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-parent:2.7.14")

	t.Run("parent has no Go module equivalent", func(t *testing.T) {
		assert.Empty(t, dep.GoModule,
			"Spring Boot parent POM has no Go module analogue; GoModule must be empty")
	})

	t.Run("parent note references go.mod", func(t *testing.T) {
		assert.Contains(t, dep.Note, "go.mod",
			"parent note must mention go.mod as the Go equivalent for version management")
	})
}

// TestDependencyMapping_DevTools validates the dev-tools entry.
func TestDependencyMapping_DevTools(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-devtools")

	t.Run("devtools has no runtime Go module", func(t *testing.T) {
		assert.Empty(t, dep.GoModule,
			"spring-boot-devtools is a dev-only tool; GoModule must be empty")
	})

	t.Run("devtools note identifies it as non-runtime", func(t *testing.T) {
		assert.NotEmpty(t, dep.Note)
	})
}

// TestDependencyMapping_DataJpa validates the JPA → database/sql mapping.
func TestDependencyMapping_DataJpa(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-data-jpa")

	t.Run("JPA maps to stdlib database/sql", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "database/sql",
			"JPA must map to the stdlib database/sql package")
	})

	t.Run("JPA note mentions PostgreSQL queries", func(t *testing.T) {
		assert.Contains(t, dep.Note, "PostgreSQL",
			"JPA note must clarify that raw PostgreSQL queries replace JPA")
	})
}

// TestDependencyMapping_Validation validates the validation → validator mapping.
func TestDependencyMapping_Validation(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-validation")

	t.Run("validation maps to go-playground/validator", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "github.com/go-playground/validator/v10",
			"validation must map to github.com/go-playground/validator/v10")
	})
}

// TestDependencyMapping_Lombok validates the Lombok → (nothing) mapping.
func TestDependencyMapping_Lombok(t *testing.T) {
	dep := findByArtifact(t, "org.projectlombok:lombok")

	t.Run("Lombok has no Go module equivalent", func(t *testing.T) {
		assert.Empty(t, dep.GoModule,
			"Lombok is compile-time boilerplate only; GoModule must be empty in Go")
	})

	t.Run("Lombok note explains it is unnecessary in Go", func(t *testing.T) {
		assert.NotEmpty(t, dep.Note,
			"Lombok note must explain why it has no Go equivalent")
	})
}

// TestDependencyMapping_Thymeleaf validates the Thymeleaf → html/template mapping.
func TestDependencyMapping_Thymeleaf(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-thymeleaf")

	t.Run("Thymeleaf maps to stdlib html/template", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "html/template",
			"Thymeleaf must map to the stdlib html/template package")
	})
}

// TestDependencyMapping_Web validates the Spring Web → net/http + chi mapping.
func TestDependencyMapping_Web(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-web")

	t.Run("web maps to stdlib net/http", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "net/http",
			"spring-boot-starter-web must map to net/http")
	})

	t.Run("web maps to chi router", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "github.com/go-chi/chi/v5",
			"spring-boot-starter-web must reference chi for idiomatic routing")
	})
}

// TestDependencyMapping_MySQLConnector_MigratedToPostgres validates the critical
// MySQL → PostgreSQL migration note.
func TestDependencyMapping_MySQLConnector_MigratedToPostgres(t *testing.T) {
	dep := findByArtifact(t, "com.mysql:mysql-connector-j")

	tests := []struct {
		name    string
		check   func(t *testing.T, dep internal.DependencyMapping)
	}{
		{
			name: "MySQL connector maps to pgx or lib/pq",
			check: func(t *testing.T, dep internal.DependencyMapping) {
				hasPgx := assert.Contains(t, dep.GoModule, "pgx") ||
					assert.Contains(t, dep.GoModule, "lib/pq")
				_ = hasPgx
			},
		},
		{
			name: "MySQL migration note must warn about PostgreSQL target",
			check: func(t *testing.T, dep internal.DependencyMapping) {
				assert.Contains(t, dep.Note, "PostgreSQL",
					"note must explicitly warn that the target DB is PostgreSQL, not MySQL")
			},
		},
		{
			name: "MySQL migration note must warn NOT to use MySQL driver",
			check: func(t *testing.T, dep internal.DependencyMapping) {
				assert.Contains(t, dep.Note, "MySQL",
					"note must reference MySQL to warn against using a MySQL driver")
			},
		},
		{
			name: "MySQL migration note contains MIGRATION_NOTE marker",
			check: func(t *testing.T, dep internal.DependencyMapping) {
				assert.Contains(t, dep.Note, "MIGRATION_NOTE",
					"MySQL connector note must carry a MIGRATION_NOTE warning marker")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t, dep)
		})
	}
}

// TestDependencyMapping_Test validates the test dependency mapping.
func TestDependencyMapping_Test(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-starter-test")

	t.Run("test dependency maps to stdlib testing package", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "testing",
			"spring-boot-starter-test must map to the stdlib testing package")
	})

	t.Run("test dependency maps to testify", func(t *testing.T) {
		assert.Contains(t, dep.GoModule, "testify",
			"spring-boot-starter-test must reference testify for assertions/mocks")
	})
}

// TestDependencyMapping_MavenPlugin validates the Maven plugin entry.
func TestDependencyMapping_MavenPlugin(t *testing.T) {
	dep := findByArtifact(t, "org.springframework.boot:spring-boot-maven-plugin")

	t.Run("spring-boot-maven-plugin has no Go module equivalent", func(t *testing.T) {
		assert.Empty(t, dep.GoModule,
			"Fat-JAR packaging is native to `go build`; GoModule must be empty")
	})

	t.Run("plugin note mentions go build", func(t *testing.T) {
		assert.Contains(t, dep.Note, "go build",
			"plugin note must explain that `go build` replaces the fat-JAR plugin")
	})
}

// TestDependencyMapping_UniqueArtifacts ensures no duplicate MavenArtifact keys
// exist in the mapping slice.
func TestDependencyMapping_UniqueArtifacts(t *testing.T) {
	seen := make(map[string]int)
	for i, dep := range internal.MavenToGoDependencies {
		if prev, exists := seen[dep.MavenArtifact]; exists {
			t.Errorf("duplicate MavenArtifact %q found at indices %d and %d",
				dep.MavenArtifact, prev, i)
		}
		seen[dep.MavenArtifact] = i
	}
}

// TestDependencyMapping_StdlibEntriesHaveNoExternalModule checks that entries
// mapped to the standard library do not accidentally reference an external module.
func TestDependencyMapping_StdlibEntriesHaveNoExternalModule(t *testing.T) {
	stdlibOnly := map[string]bool{
		"org.springframework.boot:spring-boot-starter-parent:2.7.14": true,
		"org.springframework.boot:spring-boot-devtools":               true,
		"org.projectlombok:lombok":                                    true,
		"org.springframework.boot:spring-boot-maven-plugin":           true,
	}

	for _, dep := range internal.MavenToGoDependencies {
		dep := dep
		if stdlibOnly[dep.MavenArtifact] {
			t.Run("no external module for "+dep.MavenArtifact, func(t *testing.T) {
				assert.Empty(t, dep.GoModule,
					"%s must not reference an external Go module; it has no Go equivalent",
					dep.MavenArtifact)
			})
		}
	}
}

// TestGoLanguageVersion_PrefixedWithGo ensures the constant starts with "go".
func TestGoLanguageVersion_PrefixedWithGo(t *testing.T) {
	tests := []struct {
		name    string
		version string
		prefix  string
		valid   bool
	}{
		{
			name:    "current value has go prefix",
			version: internal.GoLanguageVersion,
			prefix:  "go",
			valid:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.valid {
				assert.True(t,
					len(tc.version) >= 2 && tc.version[:2] == tc.prefix,
					"GoLanguageVersion %q must start with %q", tc.version, tc.prefix)
			}
		})
	}
}

// TestModulePath_IsValidImportPath performs a basic structural check.
func TestModulePath_IsValidImportPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		mustStart string
	}{
		{
			name:      "module path starts with github.com",
			path:      internal.ModulePath,
			mustStart: "github.com/",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc