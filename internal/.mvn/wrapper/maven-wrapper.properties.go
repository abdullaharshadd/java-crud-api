// Package wrapper documents the migration of the Maven Wrapper
// configuration (.mvn/wrapper/maven-wrapper.properties) to the Go toolchain.
//
// MIGRATION_NOTE: maven-wrapper.properties is a Maven Wrapper descriptor,
// not application logic. It pins the Maven distribution and wrapper JAR
// versions so every developer/CI runner builds with an identical Maven,
// guaranteeing reproducible Java builds. There is no line of executable
// behavior to port — nothing in it maps to a runtime code path.
//
// Go has no direct in-source analogue. The equivalent "reproducible build
// tooling" responsibilities are handled by:
//
//   - The `go` directive + `toolchain` line in go.mod
//         -> pins the Go language/toolchain version, replacing the pinned
//            Maven distribution (distributionUrl). Since Go 1.21 the
//            `toolchain` directive lets `go` auto-download and use an exact
//            toolchain version, mirroring what mvnw does for Maven.
//   - go.mod / go.sum
//         -> pin and checksum-verify every dependency version, replacing
//            Maven's role of resolving artifacts (wrapperUrl fetched the
//            wrapper JAR that did that resolution in Java).
//   - The absence of a wrapper script
//         -> Go's single self-contained toolchain removes the need for a
//            mvnw/mvnw.cmd bootstrap script and its wrapper JAR entirely.
//
// The original file's data, retained here for reference only:
//
//   distributionUrl = https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip
//   wrapperUrl      = https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar
//
// i.e. Apache Maven 3.8.7 with Maven Wrapper 3.1.1. These version pins are
// no longer meaningful once the project is a Go module; capture the target
// Go toolchain version in go.mod instead.
//
// No exported symbols are declared: there is genuinely no behavior to
// migrate, and inventing one would be misleading. This file exists purely
// to record the migration decision.
package wrapper
