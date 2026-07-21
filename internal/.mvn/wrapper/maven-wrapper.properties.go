// Package wrapper documents the migration status of the Maven Wrapper
// configuration file (.mvn/wrapper/maven-wrapper.properties).
//
// MIGRATION_NOTE: The source file is NOT Java source code, Spring
// configuration, or any runtime-bearing artifact. It is a build-tool
// bootstrap properties file consumed by the Maven Wrapper scripts
// (mvnw / mvnw.cmd) to download a pinned version of Maven before a
// JVM build runs.
//
// There is no business logic, HTTP route, Spring bean, or runtime
// behavior to migrate. In a Go project the entire Maven build-tooling
// layer is replaced by the Go toolchain:
//
//   - Dependency and version pinning  -> go.mod / go.sum
//   - Reproducible builds             -> Go modules + GOFLAGS=-mod=readonly
//   - Toolchain version pinning       -> the `go`/`toolchain` directives in go.mod
//   - Task running (mvn goals)        -> Makefile, Taskfile, or `go run` targets
//
// ACTION REQUIRED (manual): This file, along with the .mvn/ directory,
// mvnw, mvnw.cmd, and pom.xml, should be DELETED from a pure-Go
// repository. They serve no purpose outside a Maven/JVM build.
//
// The pinned versions from the original file are preserved here only as
// historical reference for whoever performs the cleanup:
//
//   distributionUrl = https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip
//   wrapperUrl      = https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar
//
// This .go file intentionally contains no executable code.
package wrapper
