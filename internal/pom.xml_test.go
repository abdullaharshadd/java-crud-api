```go
package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"smartcontact/internal"
)

func TestMavenProjectCoordinates(t *testing.T) {
	tests := []struct {
		name             string
		scenario         string
		constant         string
		expectedValue    string
		invariantMessage string
	}{
		{
			name:             "groupId matches expected Maven coordinate",
			scenario:         "build system resolves project identity",
			constant:         "MavenGroupID",
			expectedValue:    "com.smartContact",
			invariantMessage: "artifact must be uniquely identified by groupId:artifactId:version",
		},
		{
			name:             "artifactId matches expected Maven coordinate",
			scenario:         "build system resolves project identity",
			constant:         "MavenArtifactID",
			expectedValue:    "SmartContact",
			invariantMessage: "artifact must be uniquely identified by groupId:artifactId:version",
		},
		{
			name:             "version matches expected SNAPSHOT version",
			scenario:         "build system resolves project identity",
			constant:         "MavenVersion",
			expectedValue:    "0.0.1-SNAPSHOT",
			invariantMessage: "artifact must be uniquely identified by groupId:artifactId:version",
		},
		{
			name:             "project name matches expected human-readable name",
			scenario:         "build system resolves project identity",
			constant:         "ProjectName",
			expectedValue:    "SmartContact",
			invariantMessage: "artifact must be uniquely identified by groupId:artifactId:version",
		},
		{
			name:             "project description matches expected description",
			scenario:         "build system resolves project identity",
			constant:         "ProjectDescription",
			expectedValue:    "smart Contact project",
			invariantMessage: "artifact must be uniquely identified by groupId:artifactId:version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.constant {
			case "MavenGroupID":
				assert.Equal(t, tt.expectedValue, internal.MavenGroupID,
					"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			case "MavenArtifactID":
				assert.Equal(t, tt.expectedValue, internal.MavenArtifactID,
					"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			case "MavenVersion":
				assert.Equal(t, tt.expectedValue, internal.MavenVersion,
					"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			case "ProjectName":
				assert.Equal(t, tt.expectedValue, internal.ProjectName,
					"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			case "ProjectDescription":
				assert.Equal(t, tt.expectedValue, internal.ProjectDescription,
					"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			default:
				t.Fatalf("unknown constant: %s", tt.constant)
			}
		})
	}
}

func TestMavenCoordinatesUniqueIdentity(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "groupId is non-empty for unique identification",
			scenario: "build system resolves project identity",
			check: func(t *testing.T) {
				assert.NotEmpty(t, internal.MavenGroupID, "groupId must not be empty for unique identification")
			},
		},
		{
			name:     "artifactId is non-empty for unique identification",
			scenario: "build system resolves project identity",
			check: func(t *testing.T) {
				assert.NotEmpty(t, internal.MavenArtifactID, "artifactId must not be empty for unique identification")
			},
		},
		{
			name:     "version is non-empty for unique identification",
			scenario: "build system resolves project identity",
			check: func(t *testing.T) {
				assert.NotEmpty(t, internal.MavenVersion, "version must not be empty for unique identification")
			},
		},
		{
			name:     "combined coordinates produce unique identity string",
			scenario: "build system resolves project identity",
			check: func(t *testing.T) {
				combined := internal.MavenGroupID + ":" + internal.MavenArtifactID + ":" + internal.MavenVersion
				expected := "com.smartContact:SmartContact:0.0.1-SNAPSHOT"
				assert.Equal(t, expected, combined,
					"fully qualified Maven coordinates must match expected GAV triple")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestSpringBootParentInheritance(t *testing.T) {
	tests := []struct {
		name             string
		scenario         string
		expectedVersion  string
		invariantMessage string
	}{
		{
			name:             "spring boot version is 2.7.14 as declared in parent",
			scenario:         "resolving managed dependency versions",
			expectedVersion:  "2.7.14",
			invariantMessage: "parent version must be 2.7.14",
		},
		{
			name:             "spring boot version is non-empty",
			scenario:         "resolving managed dependency versions",
			expectedVersion:  "2.7.14",
			invariantMessage: "all Spring Boot managed dependencies align with the 2.7.14 release train",
		},
		{
			name:             "spring boot version contains major.minor.patch segments",
			scenario:         "resolving managed dependency versions",
			expectedVersion:  "2.7.14",
			invariantMessage: "parent version must be 2.7.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedVersion, internal.SpringBootVersion,
				"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
		})
	}
}

func TestSpringBootVersionFormat(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "spring boot version major segment is 2",
			scenario: "resolving managed dependency versions",
			check: func(t *testing.T) {
				assert.Equal(t, "2.7.14", internal.SpringBootVersion)
				assert.Contains(t, internal.SpringBootVersion, "2.",
					"spring boot version must be from the 2.x line")
			},
		},
		{
			name:     "spring boot version minor segment is 7",
			scenario: "resolving managed dependency versions",
			check: func(t *testing.T) {
				assert.Contains(t, internal.SpringBootVersion, ".7.",
					"spring boot version must be 2.7.x")
			},
		},
		{
			name:     "spring boot version patch segment is 14",
			scenario: "resolving managed dependency versions",
			check: func(t *testing.T) {
				assert.Contains(t, internal.SpringBootVersion, ".14",
					"spring boot version patch must be 14")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestJavaVersionConfiguration(t *testing.T) {
	tests := []struct {
		name             string
		scenario         string
		expectedVersion  string
		invariantMessage string
	}{
		{
			name:             "java version property equals 17",
			scenario:         "compiling sources",
			expectedVersion:  "17",
			invariantMessage: "java.version property must equal 17",
		},
		{
			name:             "java version is non-empty",
			scenario:         "compiling sources",
			expectedVersion:  "17",
			invariantMessage: "java.version property must equal 17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedVersion, internal.JavaVersion,
				"scenario: %s, invariant: %s", tt.scenario, tt.invariantMessage)
			assert.NotEmpty(t, internal.JavaVersion,
				"JavaVersion must not be empty for build configuration")
		})
	}
}

func TestJavaVersionIs17(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "java version is exactly the string '17'",
			scenario: "compiling sources",
			check: func(t *testing.T) {
				assert.Equal(t, "17", internal.JavaVersion,
					"java.version property must equal 17 exactly")
			},
		},
		{
			name:     "java version is not an older LTS version",
			scenario: "compiling sources",
			check: func(t *testing.T) {
				assert.NotEqual(t, "11", internal.JavaVersion,
					"java version must not be 11")
				assert.NotEqual(t, "8", internal.JavaVersion,
					"java version must not be 8")
				assert.NotEqual(t, "1.8", internal.JavaVersion,
					"java version must not be 1.8")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestRuntimeDependencySetDocumentation(t *testing.T) {
	// These tests validate that the constants correctly document the original
	// dependency intent as described in the behavioral specs.
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "spring boot version documents the web starter parent version",
			scenario: "assembling runtime classpath",
			check: func(t *testing.T) {
				// spring-boot-starter-web is governed by parent 2.7.14
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"web starter embedded server version is governed by 2.7.14 parent")
			},
		},
		{
			name:     "spring boot version documents JPA/Hibernate parent version",
			scenario: "assembling runtime classpath",
			check: func(t *testing.T) {
				// spring-boot-starter-data-jpa is governed by parent 2.7.14
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"JPA/Hibernate version is governed by 2.7.14 parent")
			},
		},
		{
			name:     "spring boot version documents Thymeleaf template engine parent version",
			scenario: "assembling runtime classpath",
			check: func(t *testing.T) {
				// spring-boot-starter-thymeleaf is governed by parent 2.7.14
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"Thymeleaf version is governed by 2.7.14 parent")
			},
		},
		{
			name:     "spring boot version documents validation starter parent version",
			scenario: "assembling runtime classpath",
			check: func(t *testing.T) {
				// spring-boot-starter-validation is governed by parent 2.7.14
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"validation starter version is governed by 2.7.14 parent")
			},
		},
		{
			name:     "project name documents the packaged artifact name",
			scenario: "packaging final artifact",
			check: func(t *testing.T) {
				assert.Equal(t, "SmartContact", internal.ProjectName,
					"packaged jar must be named after the project")
			},
		},
		{
			name:     "maven artifact id documents the final jar artifact id",
			scenario: "packaging final artifact",
			check: func(t *testing.T) {
				assert.Equal(t, "SmartContact", internal.MavenArtifactID,
					"artifact id must match SmartContact for the produced jar")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestTestDependencySetDocumentation(t *testing.T) {
	// The behavioral spec states spring-boot-starter-test must be test scope.
	// In Go terms, we validate that constants correctly capture original metadata
	// such that test-scope dependencies are not confused with runtime ones.
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "spring boot version referenced for test starter alignment",
			scenario: "running tests",
			check: func(t *testing.T) {
				// spring-boot-starter-test aligns with parent 2.7.14
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"test starter version must align with 2.7.14 parent")
			},
		},
		{
			name:     "project name is non-empty for test artifact identification",
			scenario: "running tests",
			check: func(t *testing.T) {
				assert.NotEmpty(t, internal.ProjectName,
					"project name must be defined for test context identification")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestSpringBootMavenPluginBuildDocumentation(t *testing.T) {
	// The behavioral spec states the spring-boot-maven-plugin produces an
	// executable jar with lombok excluded. In Go terms we validate the constants
	// that document this intent.
	tests := []struct {
		name     string
		scenario string
		check    func(t *testing.T)
	}{
		{
			name:     "project is identified as SmartContact for executable jar naming",
			scenario: "packaging the application",
			check: func(t *testing.T) {
				assert.Equal(t, "SmartContact", internal.ProjectName,
					"executable jar must be named SmartContact")
				assert.Equal(t, "SmartContact", internal.MavenArtifactID,
					"artifact id must be SmartContact for the repackaged executable jar")
			},
		},
		{
			name:     "spring boot version governs plugin and repackaging behavior",
			scenario: "packaging the application",
			check: func(t *testing.T) {
				assert.Equal(t, "2.7.14", internal.SpringBootVersion,
					"spring-boot-maven-plugin version is derived from 2.7.14 parent")
			},
		},
		{
			name:     "version includes SNAPSHOT to indicate development artifact",
			scenario: "packaging the application",
			check: func(t *testing.T) {
				assert.Contains(t, internal.MavenVersion, "SNAPSHOT",
					"artifact version must contain SNAPSHOT for development builds")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestGlobalInvariants(t *testing.T) {
	tests := []struct {
		name      string
		invariant string
		check     func(