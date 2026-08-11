// Package wrapper documents the migration of the Maven Wrapper
// configuration (.mvn/wrapper/maven-wrapper.properties) to the Go
// ecosystem.
//
// MIGRATION_NOTE: maven-wrapper.properties is a build-tooling configuration
// file, not source code. It pins the exact versions of Apache Maven and the
// Maven Wrapper that the ./mvnw script downloads, guaranteeing reproducible
// Java builds across developer machines and CI without a pre-installed
// Maven. It has no runtime behavior to translate and there is no idiomatic
// Go "file" equivalent.
//
// The Go toolchain solves the same reproducibility problem differently:
//
//   - The build tool itself is pinned via the `go` and `toolchain` directives
//     in go.mod (e.g. `go 1.22`), and the `go` command transparently fetches
//     the matching toolchain when needed. There is no separate wrapper
//     script or JAR to download.
//   - Dependency versions and their checksums are pinned in go.mod / go.sum,
//     which replace Maven's coordinate/version pinning.
//
// The original source declared two coordinates, recorded here for reference:
//
//   distributionUrl -> apache-maven 3.8.7
//                      (the Maven build engine version used for the Java build)
//   wrapperUrl      -> maven-wrapper 3.1.1
//                      (the bootstrap wrapper that downloads Maven)
//
// Neither maps to a runtime Go dependency; both are Java build-time tooling
// and are intentionally discarded. The real, authoritative version pinning
// for this Go project lives in go.mod and go.sum at the repository root.
//
// This file declares no runtime types or functions, so it cannot collide
// with any other migrated symbol in the module.
package wrapper
