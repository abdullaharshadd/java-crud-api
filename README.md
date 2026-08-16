```markdown
# java-crud-api (Go Migration)

A RESTful CRUD API for managing smart contacts. Originally implemented with Java and Spring Boot, this project has been migrated to Go using the standard library. The API provides endpoints for creating, reading, updating, and deleting user/contact records, with structured error handling and service-layer separation.

> ⚠️ **Migration Warning:** Overall migration confidence is **0%**. Every file in this project requires manual review before this code is considered production-ready. Do not run this in production without a thorough audit.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (standard library) |
| HTTP Router | `net/http` |
| Database | (see Migration Notes — not yet configured) |
| Testing | `testing` package (standard library) |
| Build | `go.mod` / `go build` |

---

## Prerequisites

- Go 1.21 or later
- A supported relational database (PostgreSQL recommended — see Migration Notes)
- Node.js / npm (detected in setup plan — verify whether this is actually required)
- Git

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install Dependencies

The setup plan detected the following install command:

```bash
npm install
```

> ⚠️ **Review required:** `npm install` is inconsistent with a Go project. Verify whether a JavaScript tooling component (e.g., a frontend, code generator, or migration script) is present. For the Go application itself, dependencies are managed via `go.mod`:
>
> ```bash
> go mod tidy
> ```

### 3. Environment Setup

No environment variables were automatically detected during migration. However, the original `application.properties` file contained database connection details and server configuration. You must manually define these before running the application.

Create a `.env` file or export variables directly (see [Environment Variables](#environment-variables) below).

### 4. Database Setup

No database setup command was detected in the migration plan. The original Spring project used JPA/Hibernate which auto-managed the schema.

You must manually:
1. Provision a database instance.
2. Create the required schema/tables (no migration scripts were generated).
3. Set the connection string in your environment.

> ⚠️ See [Known Limitations](#known-limitations) for details on what was not migrated.

### 5. Run the Application

No run command was detected in the migration plan. Once the code has been reviewed and environment is configured, build and run with:

```bash
go build -o java-crud-api ./...
./java-crud-api
```

Or run directly:

```bash
go run main.go
```

> ⚠️ Verify the entry point file name. `main.go` is assumed; confirm after reviewing the migrated source.

---

## Running Tests

No test command was detected in the migration plan. Run Go tests with:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

> ⚠️ The original test files (`SmartContactApplicationTests.java`, `UserServiceImpTest.java`) were migrated with low confidence. Expect test logic to be incomplete or incorrect. See [Manual Review Required](#manual-review-required).

---

## Environment Variables

No environment variables were automatically detected. Based on the original `application.properties`, the following are likely required. **Verify and expand this table after reviewing the migrated code.**

| Variable | Description | Example |
|---|---|---|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_NAME` | Database name | `smartcontact` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `yourpassword` |
| `SERVER_PORT` | Port the HTTP server listens on | `8080` |

> ⚠️ `src/main/resources/application.properties` was flagged as a low-confidence migration. Cross-reference the original file to capture all required configuration values before running the application.

---

## Architecture Overview

The migrated Go project follows a layered structure modelled after the original Spring package layout:

```
.
├── go.mod
├── main.go                  # Application entry point, HTTP server setup
├── model/
│   └── user.go              # Data models (migrated from com.smartContact.model)
├── service/
│   └── user_service.go      # Business logic (migrated from Spring @Service layer)
├── handler/                 # HTTP handlers (migrated from Spring @RestController)
│   └── user_handler.go
├── error/
│   ├── user_not_found.go    # Custom error type (UserNotFoundException)
│   └── error_message.go     # Error response model (ErrorMessage)
├── middleware/
│   └── exception_handler.go # Global error handling (RestResponseEntityExceptionHandling)
└── test/
    ├── app_test.go           # Application-level tests
    └── user_service_test.go  # Service unit tests
```

> ⚠️ This structure is based on the original Java package layout. The actual generated file structure may differ. Verify against the repository contents.

**Key patterns:**
- Spring `@RestController` → Go `net/http` handler functions registered on a `ServeMux`
- Spring `@Service` / `@Repository` → Go structs with method receivers and interface definitions
- Spring `@ControllerAdvice` exception handling → Go middleware or centralized error-checking in handlers
- Spring JPA entities → Go structs with database/sql or equivalent

---

## Migration Notes

This section documents what changed from the original Java/Spring Boot codebase.

### Build System
- **Before:** Maven (`pom.xml`, `.mvn/wrapper/maven-wrapper.properties`)
- **After:** Go modules (`go.mod`, `go.sum`)
- The `maven-wrapper.properties` file was explicitly identified as unmigrable and has been dropped. There is no equivalent in Go.

### Dependency Injection
- Spring's `@Autowired` / `@Component` / `@Service` annotations have no direct equivalent.
- Go uses explicit constructor functions and interface-based dependency injection. Verify that all wiring is done correctly in `main.go` or an initializer function.

### Exception Handling
- `RestResponseEntityExceptionHandling.java` (a `@ControllerAdvice` class) has been migrated with **low confidence**.
- In Go, there are no exceptions. Error handling is explicit using `error` return values and custom error types.
- The migrated code likely wraps handler responses to check for known error types (e.g., user-not-found) and returns appropriate HTTP status codes.

### ORM / Database
- Spring Data JPA / Hibernate has no direct equivalent in Go's standard library.
- Database access must be implemented using `database/sql` or a third-party library (e.g., `sqlx`, `pgx`).
- **No ORM was automatically selected or configured.** Schema management (previously handled by Hibernate `ddl-auto`) must be done manually.

### Error Types
- `UserNotFoundException.java` → custom Go `error` struct.
- `ErrorMessage.java` → Go struct used for JSON error responses.
- Both were migrated with low confidence; verify that error wrapping and HTTP response serialization behave correctly.

### Testing
- JUnit / Mockito tests have been translated to Go's `testing` package.
- Spring context bootstrapping in `SmartContactApplicationTests.java` has no equivalent; application-level tests must be restructured.
- Mock implementations previously provided by Mockito must be replaced with Go interface mocks (manually written or generated with a tool like `mockery`).

### Configuration
- `application.properties` has been replaced with environment variable lookups.
- Verify all configuration keys were captured (see [Environment Variables](#environment-variables)).

---

## Known Limitations

### Unmigrable Components

| File | Reason | Action Required |
|---|---|---|
| `.mvn/wrapper/maven-wrapper.properties` | Maven build tooling with no equivalent in Go. Tied entirely to the Java/Maven ecosystem. | **Do not use.** Replace with `go.mod`. Already dropped from the migration. |

### Low-Confidence Migrations

The following files were migrated but the output is unreliable. Logic may be missing, incorrect, or structurally incompatible with Go idioms:

| File | Risk |
|---|---|
| `pom.xml` → `go.mod` | Dependency mapping may be incomplete or missing equivalents entirely. |
| `application.properties` | Configuration keys may not all be captured as environment variables. |
| `UserNotFoundException.java` | Error type and unwrapping logic must be verified. |
| `ErrorMessage.java` | JSON serialization field names and structure must be confirmed. |
| `RestResponseEntityExceptionHandling.java` | Global error handling approach is architecturally different in Go; logic likely requires a full rewrite. |
| `SmartContactApplicationTests.java` | Spring context tests cannot be directly translated; tests likely do not compile or pass. |
| `UserServiceImpTest.java` | Mockito mocks are not valid in Go; mock strategy must be reimplemented. |

---

## Manual Review Required

The following files **must be manually reviewed and verified** by a developer before this application is usable:

- [ ] `go.mod` — Confirm all required dependencies are present and versions are appropriate.
- [ ] `main.go` — Verify server setup, route registration, and dependency wiring.
- [ ] `model/user.go` — Confirm struct fields match the original Java entity, including JSON tags and DB column mappings.
- [ ] `error/user_not_found.go` — Verify the error type satisfies the `error` interface and is checked correctly in handlers.
- [ ] `error/error_message.go` — Confirm JSON response structure matches what API consumers expect.
- [ ] `middleware/exception_handler.go` — This is the highest-risk file. The `@ControllerAdvice` pattern requires a complete rethink in Go. Verify all error cases are handled and return correct HTTP status codes.
- [ ] `service/user_service.go` — Confirm business logic is complete and that database calls are correctly implemented.
- [ ] `test/app_test.go` — Rewrite or remove; Spring context tests cannot be directly translated.
- [ ] `test/user_service_test.go` — Replace Mockito mocks with Go interface mocks and verify test coverage.
- [ ] Database layer — No ORM was configured. Confirm a database access strategy has been chosen and implemented.
- [ ] Environment variables — Audit `application.properties` from the original repo and ensure all values are externalized correctly.
- [ ] `npm install` — Determine whether this command is actually required and document why if so.

---

## Original Project

- **Source:** [abdullaharshadd/java-crud-api](https://github.com/abdullaharshadd/java-crud-api)
- **Original stack:** Java 11+, Spring Boot, Spring Data JPA, Maven
- **Migration:** Java/Spring → Go/standard library
- **Modules migrated:** 14/14
- **Overall confidence:** 0% — full manual review is mandatory
```