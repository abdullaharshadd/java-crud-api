```markdown
# java-crud-api (Go Migration)

A CRUD REST API for managing smart contacts, migrated from Java/Spring Boot to Go using the standard library. The application provides user management endpoints with structured error handling.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (1.21+) |
| HTTP Router | `net/http` (standard library) |
| Database Driver | To be confirmed — see [Migration Notes](#migration-notes) |
| Testing | `testing` (standard library) |
| Build Tool | Go modules (`go.mod`) |

---

## Prerequisites

- Go 1.21 or higher
- A running database instance (see [Migration Notes](#migration-notes) — original used a Spring datasource; exact DB type must be confirmed from `application.properties` review)
- Node.js / npm (only if the detected `npm install` step applies to a front-end or tooling component — verify this is not a migration artifact)

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install dependencies

> **Note:** An `npm install` command was detected during migration analysis. Verify whether this applies to a front-end component or tooling. For the Go backend:

```bash
go mod download
```

If there is an accompanying front-end or tooling layer:

```bash
npm install
```

### 3. Configure environment

No environment variables were automatically detected during migration. However, the original `application.properties` was flagged for manual review. You will likely need to configure database connection details before running.

Copy any example config if one exists, or create the necessary configuration manually:

```bash
cp config.example.yaml config.yaml   # if applicable
```

See the [Environment Variables](#environment-variables) section.

### 4. Database setup

> **⚠️ Manual step required.** The original `application.properties` was not fully migrated (low confidence). You must inspect the original file and configure the equivalent database connection in the Go application.

Run your database migrations if a migration tool has been set up:

```bash
# Example — confirm actual migration tooling used
go run ./cmd/migrate
```

### 5. Run the application

> **⚠️ No run command was detected automatically.** Locate the `main.go` entry point and run:

```bash
go run ./cmd/server
```

Or build and run:

```bash
go build -o api ./cmd/server
./api
```

---

## Running Tests

```bash
go test ./...
```

To run with verbose output:

```bash
go test -v ./...
```

To run with coverage:

```bash
go test -cover ./...
```

> **⚠️ Note:** `UserServiceImpTest.java` was flagged as low confidence during migration. The equivalent Go tests must be manually verified. See [Manual Review Required](#manual-review-required).

---

## Environment Variables

No environment variables were automatically extracted during migration. The following are **expected** based on the original Spring configuration — confirm by reviewing `application.properties`:

| Variable | Description | Required | Default |
|---|---|---|---|
| `DB_HOST` | Database host | Likely yes | — |
| `DB_PORT` | Database port | Likely yes | — |
| `DB_NAME` | Database name | Likely yes | — |
| `DB_USER` | Database username | Likely yes | — |
| `DB_PASSWORD` | Database password | Likely yes | — |
| `SERVER_PORT` | Port the API listens on | No | `8080` |

> These variable names are inferred. Verify against the reviewed `application.properties` and update this table accordingly.

---

## Architecture Overview

The migrated Go project follows a layered structure equivalent to the original Spring architecture:

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handler/             # HTTP handlers (equiv. Spring @RestController)
│   ├── service/             # Business logic (equiv. UserService / UserServiceImp)
│   ├── model/               # Data models (equiv. com.smartContact.model)
│   │   └── error_message.go # Migrated from ErrorMessage.java
│   └── error/               # Error handling middleware
│       └── handler.go       # Migrated from RestResponseEntityExceptionHandling.java
├── go.mod
├── go.sum
└── README.md
```

**Key structural mappings:**

| Spring Component | Go Equivalent |
|---|---|
| `@RestController` | Handler functions in `internal/handler/` |
| `@Service` / `UserServiceImp` | Structs with methods in `internal/service/` |
| `@ControllerAdvice` exception handler | Middleware / error handler in `internal/error/` |
| `application.properties` | Environment variables or config file |
| Spring dependency injection | Manual constructor injection or `wire` |

---

## Migration Notes

This project was migrated from Java 11+ / Spring Boot to Go standard library. The overall migration confidence is **0%**, meaning **all migrated files require human review before the application is considered production-ready.**

### Key changes from the original Spring codebase

- **Dependency injection:** Spring's `@Autowired` and `@Service` annotations have no direct equivalent. Dependencies are wired manually via constructors or a DI helper. Verify `UserService` and `UserServiceImp` equivalents.
- **Exception handling:** Spring's `@ControllerAdvice` / `RestResponseEntityExceptionHandling` is replaced with Go middleware or wrapper functions that intercept handler errors and write structured JSON error responses.
- **Error model:** `ErrorMessage.java` is migrated to a Go struct. Verify JSON field tags match the original API contract.
- **Build system:** `pom.xml` is replaced by `go.mod`. All dependency versions must be confirmed — the original Maven dependencies have no automatic Go equivalents.
- **Application properties:** Spring's `application.properties` configuration is not natively supported in Go. Configuration must be loaded via environment variables, a config file, or a library such as `viper`. **This was not automatically migrated.**
- **Test framework:** JUnit/Mockito tests are replaced with Go's `testing` package. Mock implementations must be hand-written or generated with a tool such as `mockery`.

---

## Known Limitations

All 14 modules were processed, but **migration confidence is 0% overall**. No modules are listed as completely unmigrable, but the following limitations apply:

- **No verified runtime:** The application has not been confirmed to compile or run correctly post-migration. Manual fixes are expected.
- **Database layer unconfirmed:** The original datasource configuration in `application.properties` was not successfully migrated. The database driver, connection string format, and ORM/query layer must be implemented manually.
- **Test coverage unverified:** Go test files are scaffolded from Java tests but behavioral equivalence has not been confirmed.
- **npm install origin unclear:** A Node.js install step was detected but its purpose in a Java/Spring project is ambiguous. This may be a migration artifact or relate to a separate tooling/front-end component not covered here.

---

## Manual Review Required

The following files were flagged as low confidence and **must be manually reviewed and corrected** before the application is used:

| File (original Java path) | Reason for Review |
|---|---|
| `pom.xml` | Build configuration not migrated; Go module dependencies must be manually identified and added to `go.mod` |
| `src/main/resources/application.properties` | Database and server configuration not migrated; must be re-implemented as environment variables or config file |
| `src/main/java/com/smartContact/model/ErrorMessage.java` | Struct migration and JSON field tags must be verified against API contract |
| `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` | Error handling middleware logic must be verified; Go has no direct `@ControllerAdvice` equivalent |
| `src/main/java/com/smartContact/service/UserService.java` | Interface definition must be verified for completeness |
| `src/main/java/com/smartContact/service/UserServiceImp.java` | Service implementation logic must be verified; DI wiring is manual in Go |
| `src/test/java/com/smartContact/SmartContactApplicationTests.java` | Application context test has no direct equivalent; verify integration test setup |
| `src/test/java/com/smartContact/service/UserServiceImpTest.java` | Unit test logic and mock implementations must be verified for behavioral equivalence |

---

## Contributing

After completing manual review of the flagged components, update this README with:
- Confirmed environment variable names
- Actual run and test commands
- Database setup steps
- Any additional configuration required
```