# java-crud-api (Go port)

> ⚠️ **Status: migration incomplete. The project does not build.**
> This branch is an in-progress migration of `abdullaharshadd/java-crud-api` from Java/Spring to Go (standard library). The last build **failed**. Unit tests have **not been run**. Behavior comparison against the original Spring application found **mismatches**. The migration tool's self-assessed confidence is **50%**. This is an estimate, not a measured result. Read [Known limitations](#known-limitations) and [Manual review required](#manual-review-required) before working on this code.

## Overview

This is a CRUD REST API for managing `User` records. The original Spring application lived in the `com.smartContact` package. It used a JPA repository (`UserDao`), a service layer (`UserServiceImp`), a REST controller (`UserController`), and a global exception handler (`RestResponseEntityExceptionHandling`).

All 8 original modules have a Go counterpart. Several of them are flagged low-confidence and need manual review.

## Tech stack

| Concern | Target (this branch) |
|---|---|
| Language | Go (module defined in `go.mod`) |
| HTTP | Go standard library (`net/http`), routing in `cmd/server/router.go` |
| Persistence | `database/sql` with a hand-written repository (`internal/user/repository.go`). Check `go.mod` for the SQL driver in use. |
| Configuration | Environment variables (`internal/config/config.go`) |
| Containerization | `Dockerfile`, `docker-compose.yml` |

## Prerequisites

- A Go toolchain installed locally. The build sets `GOTOOLCHAIN=local`, so Go will not download a different toolchain. Your installed version must satisfy the `go` directive in `go.mod`.
- `git` and a C build toolchain. The install step targets Alpine Linux (`apk add --no-cache git build-base`) and assumes the source is at `/app`, as in the container image.
- A reachable SQL database for `DATABASE_URL`.

## Getting started

Run these steps in order, from a fresh clone to a running server.

> Because the build currently fails, step 1 is expected to fail until the issues under [Manual review required](#manual-review-required) are fixed.

### 1. Install dependencies and build

This command was detected for the project. It is written for an Alpine environment with the source mounted at `/app`:

```sh
apk add --no-cache git build-base && export GOTOOLCHAIN=local && cd /app && mkdir -p /app/bin && MAINFILE=$(grep -rlE --include='*.go' '^package main' . 2>/dev/null | grep -v '/vendor/' | grep -v '_test.go' | head -n1); if [ -z "$MAINFILE" ]; then echo 'No Go main package found; generating minimal server at cmd/server/main.go' >&2; mkdir -p cmd/server && printf '%s\n' 'package main' '' 'import (' '	"database/sql"' '	"net/http"' '	"os"' '' '	_ "github.com/go-sql-driver/mysql"' ')' '' 'func main() {' '	db, _ := sql.Open("mysql", os.Getenv("DATABASE_URL"))' '	port := os.Getenv("PORT")' '	if port == "" {' '		port = "8080"' '	}' '	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {' '		if db != nil && db.Ping() == nil {' '			w.Write([]byte("ok"))' '			return' '		}' '		w.Write([]byte("db unavailable"))' '	})' '	http.ListenAndServe(":"+port, nil)' '}' > cmd/server/main.go; MAINFILE=./cmd/server/main.go; fi && MAINDIR=$(cd "$(dirname "$MAINFILE")" && pwd) && MODDIR=$MAINDIR && while [ ! -f "$MODDIR/go.mod" ] && [ "$MODDIR" != "/" ]; do MODDIR=$(dirname "$MODDIR"); done && cd "$MODDIR" && go mod tidy && go mod download && CGO_ENABLED=0 go build -o /app/bin/server "$MAINDIR"
```

Notes on what this command does:

- It looks for a `package main` file. `cmd/server/main.go` already exists in this repository, so the fallback branch, which would generate a minimal `/healthz`-only server, should **not** run. If you see the message `No Go main package found; generating minimal server...`, something is wrong. Do not mistake that generated stub for the real application.
- It builds a static binary (`CGO_ENABLED=0`) to `/app/bin/server`.

### 2. Resolve module dependencies

This step ran successfully during migration. Run it from the module root (the directory containing `go.mod`):

```sh
go mod tidy
```

### 3. Configure environment

Set the variables listed in [Environment variables](#environment-variables). At minimum, set `DATABASE_URL`.

### 4. Database setup

No database setup or migration command was detected. The database and the table(s) backing `User` must exist before the server starts.

The original project may have relied on Hibernate/JPA schema generation. That does not exist in the Go port, so you must create the schema manually. Derive the table definition from `internal/user/model.go` and the SQL in `internal/user/repository.go`.

### 5. Run

```sh
/app/bin/server
```

## Running tests

```sh
GOTOOLCHAIN=local go test ./...
```

Tests have **not** been run on this branch, and the build currently fails, so expect this command to fail. No Go test files are listed in the repository, so no ported test coverage exists yet.

## Environment variables

| Variable | Purpose | Notes |
|---|---|---|
| `DATABASE_URL` | Connection string for the SQL database | Required. The format depends on the driver declared in `go.mod`. |
| `PORT` | HTTP listen port | See `internal/config/config.go` for the default, if any. |
| `JWT_SECRET` | Secret for JWT signing/verification | Detected as required. Confirm in `internal/config/config.go` whether and how it is used. The original's auth behavior was not verified. |
| `GOTOOLCHAIN` | Go toolchain selection | Set to `local` by the install and test commands to prevent toolchain auto-download. |
| `CGO_ENABLED` | cgo toggle at build time | Set to `0` by the build command to produce a static binary. |

## Architecture overview

```
cmd/server/
  main.go          Entry point: loads config, wires dependencies, starts HTTP server
  router.go        Route registration (replaces Spring @RequestMapping)
internal/config/
  config.go        Environment-based configuration (replaces application.properties)
internal/httpapi/
  handler.go       HTTP handlers for User endpoints (ported from UserController)
  errors.go        Error-to-HTTP-response mapping (ported from RestResponseEntityExceptionHandling)
internal/user/
  model.go         User type (ported from the User model)
  repository.go    Hand-written SQL data access (replaces UserDao / JpaRepository)
  service.go       Business logic (ported from UserServiceImp)
  errors.go        Domain errors used by service and mapped in httpapi/errors.go
Dockerfile
docker-compose.yml
```

Request flow: `router.go` → `httpapi.handler` → `user.service` → `user.repository` → database. Errors from the `user` package are translated into HTTP responses in `internal/httpapi/errors.go`.

The original-to-Go mapping above reflects the intended correspondence. Verify it against the code.

## Migration notes

The main differences from the Spring codebase:

- **No dependency injection container.** Spring's `@Autowired`/component scanning is replaced by explicit wiring in `cmd/server/main.go`.
- **No Spring Data JPA.** `UserDao` extended `JpaRepository`, whose implementation Spring generates at runtime. In Go, `internal/user/repository.go` contains hand-written SQL for the operations the service actually uses.
- **No Hibernate session semantics.** Updates must be explicit. There is no dirty checking, cascading, lazy loading, or first-level cache.
- **Exception handling becomes error values.** `@ControllerAdvice`/`@ExceptionHandler` in `RestResponseEntityExceptionHandling` is replaced by Go error types (`internal/user/errors.go`) and an explicit mapping in `internal/httpapi/errors.go`.
- **Routing.** Spring annotations are replaced by registrations with `net/http` in `cmd/server/router.go`.
- **Configuration.** Spring property files are replaced by environment variables read in `internal/config/config.go`.
- **Validation and serialization.** Bean Validation annotations and Jackson defaults do not carry over. JSON field names, null handling, and validation rules may differ. These are likely sources of the observed behavior mismatches.

## Known limitations

- **Build fails.** The project does not currently compile. The cause has not been documented here; start with `go build` output from the module root.
- **Behavior differs from the original.** Comparison against the Spring app found mismatches. The specific mismatched endpoints and responses are not recorded in this README.
- **Tests not run.** There are no ported tests and no test results.
- **Unmigratable: JpaRepository runtime-generated methods** (`UserDao`).
  - Go has no runtime proxy generation or query derivation from method names.
  - Every repository method must be implemented by hand in `internal/user/repository.go`.
  - Inherited methods the original may have used implicitly (save, find by ID, find all, delete, etc.) must each be confirmed present and correct.
  - `sqlc` is an option for generating type-safe query code.
- **Unmigratable: JPA persistence-context semantics** (`UserDao`).
  - Dirty checking, cascades, lazy loading, and the first-level cache came from Hibernate and have no `database/sql` equivalent.
  - Updates must be explicit SQL.
  - Persisting associated data must be done explicitly within transactions.
  - Lazy loading must be replaced by explicit queries.
- **No schema management.** No migration tool or schema creation step is included.
- **Install command is container-specific.** It assumes Alpine (`apk`) and an `/app` path. Running outside that environment requires adapting it.

## Manual review required

These originals were migrated with low confidence. Review their Go counterparts line by line against the Java source:

| Original (Java) | Go counterpart | What to check |
|---|---|---|
| `User` | `internal/user/model.go` | Field names and types, JSON tags vs. Jackson output, column mapping, validation constraints, ID generation strategy |
| `RestResponseEntityExceptionHandling` | `internal/httpapi/errors.go`, `internal/user/errors.go` | HTTP status codes, error response body shape, coverage of every exception type the original handled |
| `UserServiceImp` | `internal/user/service.go` | Business rules, not-found handling, update semantics (explicit vs. Hibernate dirty checking), transaction boundaries |
| `UserController` | `internal/httpapi/handler.go`, `cmd/server/router.go` | Paths, HTTP methods, path/query parameter parsing, request body decoding, response status codes and bodies |

Also verify:

- `UserDao` → `internal/user/repository.go`: every query the service relies on exists and returns equivalent results. See [Known limitations](#known-limitations).
- `cmd/server/main.go` is the real application entry point, not the fallback `/healthz` stub that the install command can generate.
- `JWT_SECRET` usage in `internal/config/config.go` matches the original's authentication behavior, if it had any.

Recommended order: fix the build → add tests for the four files above → re-run the behavior comparison against the Spring app.