// Package wrapper previously corresponded to the Maven Wrapper configuration
// (.mvn/wrapper/maven-wrapper.properties). Maven is a build tool for the JVM
// ecosystem and has no counterpart in a Go project.
//
// MIGRATION_NOTE: The original file only declared which Maven distribution and
// wrapper JAR to download so that the project could be built with a pinned
// Maven version. There is NO business logic to migrate.
//
// In an idiomatic Go project this build-tooling responsibility is handled by:
//   - The Go module system (go.mod / go.sum) for dependency pinning.
//   - The `go` toolchain directive in go.mod (e.g. `go 1.23`) for the language
//     version, and an optional `toolchain` directive to pin the exact Go
//     toolchain (Go 1.21+ downloads it automatically, analogous to the wrapper).
//   - A Makefile or shell scripts (./scripts) for build/test/lint automation.
//
// Recommended next steps (manual):
//   1. Delete the entire .mvn/ directory and the mvnw / mvnw.cmd scripts.
//   2. Add a Makefile with `build`, `test`, `lint`, and `run` targets.
//   3. Optionally pin the toolchain in go.mod, e.g.:
//        go 1.23
//        toolchain go1.23.0
//
// This file intentionally contains no runnable code; it exists only to document
// the deliberate decision to discard the Maven wrapper during migration.
package wrapper
