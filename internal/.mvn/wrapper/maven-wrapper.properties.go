// Package wrapper contains build-tooling configuration that was originally
// expressed as a Maven Wrapper properties file (.mvn/wrapper/maven-wrapper.properties).
//
// MIGRATION_NOTE: The source file is a Maven Wrapper configuration file, not
// application logic. It has no behavioral impact on the running Go program and
// is build-tooling only. A Go program built with the standard `go` toolchain
// (go build / go test / go modules) does not use Maven at all, so there is no
// direct functional equivalent to migrate.
//
// The original URLs are preserved below as exported constants purely for
// documentation and traceability. They are NOT consumed by any Go build step.
// If this repository still needs to build any residual JVM artifacts, keep the
// original .mvn/wrapper/maven-wrapper.properties file in place — deleting it
// would break `./mvnw`.
//
// REQUIRES MANUAL REVIEW: Decide whether the JVM/Maven build is still needed.
// If not, both the original properties file and this stub can be removed.
package wrapper

const (
	// DistributionURL is the URL from which the Maven distribution was downloaded
	// by the Maven Wrapper in the original project. Retained for documentation only.
	DistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// WrapperURL is the URL from which the Maven Wrapper JAR was downloaded in the
	// original project. Retained for documentation only.
	WrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
