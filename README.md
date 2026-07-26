# java-crud-api (Go Migration)

> **⚠️ Migration Notice:** This project was migrated from Java/Spring Boot to Go (standard library). Overall migration confidence is **0%**. All migrated code requires thorough manual review before use in any environment.

---

## Description

A CRUD REST API for managing user/contact records (SmartContact). The application exposes HTTP endpoints for creating, reading, updating, and deleting user entities, with structured error handling and a service/repository layered architecture.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (standard library) |
| HTTP Server | `net/http` |
| Routing | `net/http` (stdlib) or compatible mux |
| Database | To be confirmed — see Migration Notes |
| Testing | `testing` (stdlib) |
| Build Tool | Go modules (`go.mod`) |

---

## Prerequisites

- **Go** 1.21 or later — [Install Go](https://go.dev/dl/)
- **Node.js / npm** — Required for tooling only (see note below)
- A running database instance (PostgreSQL or MySQL — confirm from original `application.properties`)
- `git`

> **Note:** The setup plan detected `npm install` as an install command. This is likely a tooling artifact from the migration pipeline and does **not** indicate this is a Node.js project. Verify whether any npm-based tooling (e.g., code generation, linting scripts) was intentionally carried over.

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install Dependencies

```bash
# If npm tooling was intentionally migrated:
npm install

# Install Go dependencies:
go mod tidy
```

### 3. Environment Setup

Copy and populate the environment file:

```bash
cp .env.example .env
```

Edit `.env` with your values. See the [Environment Variables](#environment-variables) table below.

> **No environment variables were detected** during migration analysis. You must manually inspect the original `src/main/resources/application.properties` to identify all required configuration values (database URL, credentials, port, etc.) and define them in your Go configuration layer.

### 4. Database Setup

No automated DB setup script was detected. Before running the application:

1. Create your database manually.
2. Apply any schema migrations. The original project used Spring Data JPA (likely with auto-DDL); the Go version does **not** have ORM auto-migration. You must write and apply schema SQL manually.

```bash
# Example (adjust for your database):
psql -U youruser -d yourdb -f schema.sql
```

### 5. Run the Application

```bash
go run ./cmd/main.go
```

> **⚠️ The entry point path is unconfirmed.** Locate the `main.go` file in the migrated source tree and adjust the command accordingly.

---

## Running Tests

```bash
go test ./...
```

To run with verbose output:

```bash
go test -v ./...
```

> **⚠️ Test coverage is not guaranteed.** The original test files (`SmartContactApplicationTests.java`, `UserServiceImpTest.java`) were part of the low-confidence migration set. Migrated tests must be manually reviewed and likely rewritten.

---

## Environment Variables

> No environment variables were automatically detected during migration. The table below lists variables that **likely exist** based on the original Spring Boot project structure. You must verify each one against the original `application.properties`.

| Variable | Description | Example Value | Required |
|---|---|---|---|
| `DB_HOST` | Database host | `localhost` | Likely yes |
| `DB_PORT` | Database port | `5432` | Likely yes |
| `DB_NAME` | Database name | `smartcontact` | Likely yes |
| `DB_USER` | Database username | `admin` | Likely yes |
| `DB_PASSWORD` | Database password | `secret` | Likely yes |
| `SERVER_PORT` | HTTP server listen port | `8080` | Likely yes |

**Action required:** Open `src/main/resources/application.properties` in the original repository and map every `spring.datasource.*`, `server.*`, and custom property to a corresponding environment variable in the Go implementation.

---

## Architecture Overview

The migrated Go project follows the same layered structure as the original Spring Boot application, translated to idiomatic Go packages:

```
.
├── cmd/
│   └── main.go               # Application entry point
├── internal/
│   ├── controller/           # HTTP handlers (migrated from UserController.java)
│   ├── service/              # Business logic (migrated from UserService / UserServiceImp)
│   ├── repository/           # Data access layer (migrated from UserDao)
│   ├── model/                # Structs (migrated from User.java, ErrorMessage.java)
│   └── error/                # Error handling middleware (migrated from RestResponseEntityExceptionHandling)
├── go.mod
└── go.sum
```

> **⚠️ The actual directory layout may differ.** The structure above reflects the intended migration pattern. Verify the real file locations in the repository before relying on this diagram.

### Layer Responsibilities

| Layer | Original (Java/Spring) | Migrated (Go) |
|---|---|---|
| HTTP Routing & Handlers | `@RestController`, `@RequestMapping` | `net/http` handler functions |
| Business Logic | `@Service`, `UserServiceImp` | `service` package, plain structs + interfaces |
| Data Access | `@Repository`, `UserDao` (Spring Data JPA) | `repository` package — **ORM removed, raw SQL or driver required** |
| Error Handling | `@ControllerAdvice`, `RestResponseEntityExceptionHandling` | Middleware or helper functions returning JSON error responses |
| Dependency Injection | Spring IoC container | Manual wiring in `main.go` |

---

## Migration Notes

### What Changed from the Original Spring Codebase

| Concern | Spring Boot (Original) | Go Standard Library (Migrated) |
|---|---|---|
| Framework | Spring Boot 3.x | None — stdlib `net/http` only |
| Build tool | Maven (`pom.xml`) | Go modules (`go.mod`) |
| Dependency injection | Spring IoC / `@Autowired` | Manual constructor injection in `main.go` |
| ORM / Data access | Spring Data JPA / Hibernate | **No ORM.** Raw SQL via `database/sql` or a lightweight driver |
| Bean validation | `jakarta.validation` annotations | Manual validation or a Go validation library |
| Exception handling | `@ControllerAdvice` global handler | Per-handler error returns or HTTP middleware |
| Configuration | `application.properties` | Environment variables or a config file (not auto-configured) |
| Testing | JUnit 5, Spring Test, Mockito | Go `testing` package, manual mocks |
| JSON serialization | Jackson (auto-configured) | `encoding/json` (explicit marshal/unmarshal) |
| Application lifecycle | Spring application context | Explicit `http.ListenAndServe` in `main.go` |

### Database Schema

Spring Data JPA with `spring.jpa.hibernate.ddl-auto` may have auto-generated the schema. Go has no equivalent. You must:
1. Extract the schema from the original running database or from JPA entity definitions.
2. Write explicit `CREATE TABLE` SQL.
3. Apply it before running the Go application.

---

## Known Limitations

### Overall Migration Confidence: 0%

The automated migration tool reported **0% confidence** across the entire project. This means the translated code should be treated as a **starting point scaffold only**, not as production-ready or functionally equivalent code.

### Specific Limitations

| Issue | Detail |
|---|---|
| No ORM | Spring Data JPA repositories cannot be directly translated. `UserDao` repository methods must be manually rewritten as SQL queries. |
| No dependency injection framework | All wiring that Spring handled automatically must be done manually. Missing wiring will cause nil pointer panics at runtime. |
| Validation not migrated | Any `@Valid`, `@NotNull`, `@Size`, or similar annotations on `User.java` have no automatic equivalent in Go. |
| Error handling contract | The `@ControllerAdvice` global exception handler produces consistent error shapes. The Go equivalent must be manually implemented to match the same HTTP status codes and `ErrorMessage` JSON structure. |
| `application.properties` not parsed | Spring's property binding has no equivalent. All configuration must be manually wired to environment variables or a config struct. |
| Test suite non-functional | Both test files were low-confidence migrations. Spring-specific test annotations (`@SpringBootTest`, `@MockBean`) do not exist in Go. Tests need to be fully rewritten. |
| npm install command | An `npm install` command was detected in the setup plan with no clear justification. The purpose of any Node.js dependency must be confirmed before running it. |

---

## Manual Review Required

The following files/components **must be manually reviewed and corrected** by a developer before the application can be considered functional. All were flagged as low-confidence during migration:

| File (Original) | Reason for Review |
|---|---|
| `.mvn/wrapper/maven-wrapper.properties` | Maven-specific — confirm whether any equivalent Go tooling setup is needed |
| `pom.xml` | All Spring/Java dependencies need to be mapped to Go equivalents; verify nothing was dropped |
| `src/main/java/com/smartContact/model/User.java` | All field types, validation annotations, and JPA column mappings must be verified against the Go struct |
| `src/main/java/com/smartContact/model/ErrorMessage.java` | Confirm the JSON error response shape is preserved exactly in the Go `ErrorMessage` struct |
| `src/main/resources/application.properties` | Every property must be manually mapped to an environment variable in the Go config layer |
| `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` | HTTP status codes and error response bodies must match the original handler behavior |
| `src/main/java/com/smartContact/repository/UserDao.java` | Spring Data JPA method signatures do not translate automatically — every query method must be written as explicit SQL |
| `src/main/java/com/smartContact/service/UserService.java` | Verify interface method signatures match the Go interface definition exactly |
| `src/main/java/com/smartContact/service/UserServiceImp.java` | Business logic correctness must be validated line by line; Spring transaction semantics are not automatically preserved |
| `src/main/java/com/smartContact/Controller/UserController.java` | All route mappings, HTTP methods, request/response binding, and status codes must be verified against the original |
| `src/test/java/com/smartContact/SmartContactApplicationTests.java` | Spring Boot test context does not exist in Go — rewrite from scratch |
| `src/test/java/com/smartContact/service/UserServiceImpTest.java` | Mockito-based mocks must be replaced with Go interfaces and manual test doubles |

---

## Recommended First Steps for Maintainers

1. **Do not deploy this code.** The 0% confidence score means correctness cannot be assumed.
2. Compare every migrated Go file side-by-side with its original Java source.
3. Stand up the original Spring Boot application locally and document the exact API contract (endpoints, request bodies, response shapes, status codes).
4. Use that contract to write integration tests for the Go application before trusting any logic.
5. Manually implement database queries in the repository layer — this is the most critical missing piece.
6. Resolve the `npm install` ambiguity before running it in any environment.