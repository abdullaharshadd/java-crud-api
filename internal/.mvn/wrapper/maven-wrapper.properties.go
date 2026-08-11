// Package wrapper preserves, for reference only, the build-tooling
// requirements that were expressed in the Maven Wrapper configuration file
// (.mvn/wrapper/maven-wrapper.properties) of the original Java project.
//
// MIGRATION_NOTE: The source file is NOT executable business logic. It is a
// Maven Wrapper (mvnw) configuration file that told the Java build which Maven
// distribution and wrapper JAR to download so builds were reproducible across
// machines. Go has no direct equivalent: reproducible builds are handled by the
// Go toolchain itself (the `go` directive in go.mod, and optionally a
// `toolchain` directive / GOTOOLCHAIN env var). There is therefore no code to
// migrate. This file exists only to record the original Java/Maven version
// requirements as documented constants so the information survives the
// migration and can be consulted during a manual review of the build setup.
//
// Nothing here participates in the running application. It declares no HTTP
// routes (there were none to register) and touches no database.
package wrapper

const (
	// MavenDistributionURL is the Maven distribution the original Java project
	// pinned via distributionUrl in maven-wrapper.properties. Recorded here for
	// reference; the Go build does not use Maven.
	MavenDistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// MavenWrapperURL is the Maven Wrapper JAR the original Java project pinned
	// via wrapperUrl in maven-wrapper.properties. Recorded here for reference;
	// the Go build does not use Maven.
	MavenWrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"

	// MavenVersion is the Apache Maven version the Java project targeted.
	MavenVersion = "3.8.7"

	// MavenWrapperVersion is the maven-wrapper version the Java project targeted.
	MavenWrapperVersion = "3.1.1"
)
