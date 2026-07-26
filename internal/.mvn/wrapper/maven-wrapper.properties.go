// Package wrapper is a placeholder for the Maven Wrapper configuration.
//
// MIGRATION_NOTE: The source file .mvn/wrapper/maven-wrapper.properties is a
// Maven Wrapper build-tooling configuration. It specifies which Maven
// distribution and wrapper JAR to download so builds are reproducible across
// environments. This is a build-system artifact and has NO direct Go
// equivalent — Go uses Go modules (go.mod / go.sum) and the `go` toolchain
// directive for reproducible builds and toolchain pinning.
//
// This file exists only to satisfy the migration mapping. It contains no
// business logic and should be treated as documentation. The original
// properties are preserved below as constants for reference.
//
// REQUIRES MANUAL REVIEW:
//   - Delete this file if the project is fully migrated to Go modules.
//   - If a hybrid build (JVM + Go) is retained, keep the original
//     .mvn/wrapper/maven-wrapper.properties in place instead.
//   - To pin the Go toolchain version (the closest analogue to pinning the
//     Maven distribution), use the `toolchain` directive in go.mod, e.g.:
//         go 1.22
//         toolchain go1.22.5
package wrapper

// Reference values migrated verbatim from the original Maven Wrapper
// configuration. These are documentation-only; they are not consumed by any
// Go code.
const (
	// DistributionURL is the Maven distribution the wrapper would download.
	// MIGRATION_NOTE: No Go equivalent — informational only.
	DistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// WrapperURL is the Maven wrapper JAR the wrapper would download.
	// MIGRATION_NOTE: No Go equivalent — informational only.
	WrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
