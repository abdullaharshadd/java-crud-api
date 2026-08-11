// Package wrapper documents the original Maven Wrapper configuration that was
// used to bootstrap the source project's build environment.
//
// MIGRATION_NOTE: The source file (.mvn/wrapper/maven-wrapper.properties) is a
// Java/Maven build-tooling artifact. It has NO runtime behavior and NO Go
// equivalent: Go projects use the `go` toolchain and Go modules (see go.mod)
// rather than the Maven Wrapper to pin a build environment. Nothing here is
// executed by the migrated application.
//
// This file exists only to preserve the build-context metadata (Maven and
// wrapper versions) for humans auditing the migration. The Go build is driven
// entirely by go.mod / the `go` command; delete this file once the migration
// has been reviewed if you do not want the historical record in-tree.
package wrapper

// Build-context constants preserved verbatim from the original
// .mvn/wrapper/maven-wrapper.properties file. These are informational only —
// they are not consumed by any Go code and are retained solely for traceability
// (e.g. confirming the Java/Spring Boot toolchain the source was built against).
const (
	// MavenDistributionURL is the Apache Maven distribution that the original
	// project's wrapper downloaded to guarantee a consistent build. Maven 3.8.7.
	MavenDistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// MavenWrapperURL is the maven-wrapper JAR (3.1.1) used by the original
	// project to bootstrap Maven without a pre-installed distribution.
	MavenWrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
