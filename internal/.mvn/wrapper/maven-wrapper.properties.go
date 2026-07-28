// Package wrapper documents the Maven Wrapper configuration that was present
// in the original Java project. It contains no executable Go code.
//
// MIGRATION_NOTE: The source file (.mvn/wrapper/maven-wrapper.properties) is a
// Maven Wrapper properties file. Its sole purpose was to pin the exact Maven
// distribution and the Maven Wrapper JAR used to bootstrap a reproducible build
// environment, so a developer could run ./mvnw without having Maven installed
// globally. It is pure build-tooling metadata and contains no business logic.
//
// MIGRATION_NOTE: Go has no direct equivalent to the Maven Wrapper. Build
// reproducibility in Go is achieved by a completely different mechanism:
//   - The Go toolchain version is pinned via the `go` and `toolchain`
//     directives in go.mod (the analogue of the pinned distributionUrl below).
//   - Dependency versions are pinned via go.mod / go.sum (verified checksums),
//     which is the analogue of the wrapper JAR guaranteeing a consistent
//     bootstrap.
//   - The `go` command itself is the build driver; there is no separate
//     wrapper script or wrapper JAR to download (the analogue of wrapperUrl).
//
// MIGRATION_NOTE: Therefore there is no migration action required for this
// file. The mappings below are advisory only, recorded so a human reviewer can
// reproduce the original build-bootstrap intent within the Go ecosystem.
//
// Original Maven Wrapper settings (for reference):
//   distributionUrl = https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip
//     -> Go equivalent: pin the toolchain in go.mod, e.g.
//          go 1.22
//          toolchain go1.22.0
//   wrapperUrl = https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar
//     -> Go equivalent: none needed; the installed `go` command is the build
//        driver, and go.sum provides the checksum-verified bootstrap guarantee.
package wrapper
