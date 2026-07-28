// Package wrapper documents the Maven Wrapper configuration that was
// previously expressed as .mvn/wrapper/maven-wrapper.properties for the
// SmartContact Spring Boot application.
//
// MIGRATION_NOTE: maven-wrapper.properties is NOT source code and contains no
// runtime business logic to translate. It is a build-tooling reference file
// consumed by the Maven Wrapper (mvnw / mvnw.cmd) scripts. Its sole purpose is
// to pin the exact Maven distribution and wrapper JAR versions that the wrapper
// downloads on first use, guaranteeing reproducible builds across developer
// machines and CI environments regardless of any locally installed Maven.
//
// Go has no direct equivalent because the Go toolchain solves the same problem
// differently:
//
//   1. The Go compiler/toolchain version is pinned via the `go` directive in
//      go.mod (and optionally the `toolchain` directive on Go 1.21+), so the
//      exact toolchain can be fetched automatically — this replaces the
//      distributionUrl pinning below.
//   2. Reproducible dependency resolution is provided by go.mod + go.sum (with
//      cryptographic checksums), replacing the need for a downloaded wrapper
//      JAR (wrapperUrl below).
//   3. There is no separate bootstrap-launcher artifact to download; `go build`
//      and `go test` are self-contained.
//
// The constants below are preserved purely as documentation of the original
// pinned versions for historical/audit purposes. They are not consumed by any
// Go build step and require no manual review.
package wrapper

const (
	// DistributionURL is the Maven distribution archive that the original
	// Maven Wrapper downloaded and used to build the project. Preserved for
	// historical reference only; the Go toolchain version is pinned in go.mod
	// instead.
	DistributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"

	// WrapperURL is the Maven Wrapper JAR that bootstrapped the Maven download.
	// Preserved for historical reference only; Go has no equivalent bootstrap
	// artifact — go.sum provides reproducible, checksum-verified dependencies.
	WrapperURL = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)
