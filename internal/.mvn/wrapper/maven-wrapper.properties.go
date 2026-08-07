// Package wrapper documents the project's Maven Wrapper configuration.
//
// MIGRATION_NOTE: The source file was .mvn/wrapper/maven-wrapper.properties, a
// Maven Wrapper configuration file. Its sole purpose was to pin the exact
// Maven distribution and wrapper JAR versions/URLs that the ./mvnw script
// downloads so that every developer and CI machine builds the project with a
// consistent Maven toolchain. It contains no business logic and no executable
// code; it is pure build-tooling metadata consumed only by the Maven Wrapper
// bootstrap script.
//
// Go has no equivalent concept. The Go toolchain does not require a
// per-project "wrapper" that downloads a build tool, because:
//
//   - The `go` directive in go.mod pins the minimum/target Go language
//     version, and modern Go toolchains (>= 1.21) will automatically download
//     and use the exact toolchain named there (e.g. `toolchain go1.22.0`),
//     which is the closest analogue to distributionUrl pinning a Maven
//     version.
//   - go.sum cryptographically pins every dependency, replacing the need to
//     pin a wrapper JAR download URL.
//
// Therefore there is no runtime Go artifact to generate from this file. This
// source only documents the build-tooling versions that were in effect for
// the original project, for historical/reference purposes. The constants below
// preserve the exact version and URL values from the original properties file
// so the migration is lossless and greppable.
package wrapper

const (
	// MavenDistributionVersion is the Apache Maven distribution version that
	// the original project pinned via the maven-wrapper distributionUrl.
	//
	// MIGRATION_NOTE: In Go this role is played by the `go`/`toolchain`
	// directives in go.mod, not by a downloaded distribution.
	MavenDistributionVersion = "3.8.7"

	// MavenDistributionURL is the exact distributionUrl the Maven Wrapper used
	// to fetch the pinned Maven distribution.
	MavenDistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// MavenWrapperVersion is the version of the maven-wrapper JAR the original
	// project pinned via the maven-wrapper wrapperUrl.
	MavenWrapperVersion = "3.1.1"

	// MavenWrapperURL is the exact wrapperUrl the Maven Wrapper used to
	// bootstrap itself.
	MavenWrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
