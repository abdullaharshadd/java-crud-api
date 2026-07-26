// Package wrapper documents the Maven Wrapper configuration that was used
// by the original Java/Maven project. This file has NO runtime behavior in
// the Go port; it exists purely to preserve build-parity metadata so that
// anyone reproducing the original build knows exactly which Maven and Maven
// Wrapper versions were pinned.
//
// MIGRATION_NOTE: The source artifact was `.mvn/wrapper/maven-wrapper.properties`,
// a Maven Wrapper bootstrap configuration. Go projects do not use Maven; the
// Go toolchain and modules (go.mod / go.sum) fully replace this mechanism.
// There is no idiomatic Go equivalent, so the original values are captured
// here as exported constants for documentation and audit purposes only.
// This file can be safely deleted once build-parity notes have been recorded.
//
// MIGRATION_NOTE: No HTTP routes were present in the source file, so none are
// registered here despite the generic routing requirement.
package wrapper

// Maven build-parity constants preserved from the original
// `.mvn/wrapper/maven-wrapper.properties`. These are informational only and
// are not consumed by any Go code path.
const (
	// MavenDistributionVersion is the pinned Apache Maven distribution version
	// used by the original project's build.
	MavenDistributionVersion = "3.8.7"

	// MavenWrapperVersion is the pinned Maven Wrapper version used to bootstrap
	// the original project's build.
	MavenWrapperVersion = "3.1.1"

	// DistributionURL is the URL from which the Apache Maven distribution was
	// downloaded by the Maven Wrapper.
	DistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// WrapperURL is the URL from which the Maven Wrapper JAR was downloaded.
	WrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
