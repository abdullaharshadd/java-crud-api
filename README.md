# java-crud-api (Go Migration)

> **⚠️ Migration Warning:** This project was automatically migrated from Java/Spring Boot to Go/standard library with **0% overall confidence**. Every file requires manual review before this code is considered production-ready.

---

## Description

A CRUD API for managing smart contacts (`SmartContact`). The application exposes REST endpoints for user/contact management, with error handling middleware and a service layer. Originally built with Spring Boot; this repository contains the migrated Go equivalent.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (1.21+) |
| HTTP Server | `net/http` (standard library) |
| Routing | `net/http` ServeMux |
| Database | See [Migration Notes](#migration-notes) |
| Testing | `testing` (standard library) |
| Build/Deps | Go Modules (`go.mod`) |

---

## Prerequisites

- Go 1.21 or higher — [install guide](https://go.dev/doc/install)
- Node.js / npm (detected in setup commands — see [Migration Notes](#migration-notes))
- A running database instance (see [Database Setup](#database-setup))
- `git`

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install dependencies

> **⚠️ Note:** The setup plan detected `npm install` as the install command. This is likely a migration artifact and does not reflect a typical Go project. Verify whether a `package.json` is intentionally present.

```bash
# If a package.json exists for tooling (e.g., code generation, linting):
npm install

# Install Go module dependencies:
go mod tidy
```

### 3. Environment setup

No environment variables were automatically detected during migration. However, the original `application.properties` was migrated with low confidence. You must manually verify what configuration values are required.

Check the migrated config file and set variables as needed:

```bash
cp .env.example .env   # if an example file exists
# Edit .env with your values
```

See the [Environment Variables](#environment-variables) section for details.

### 4. Database setup

> **⚠️ No database setup command was detected during migration.** The original Spring project used `application.properties` for datasource configuration. That file was migrated with **low confidence** — the database driver, connection string, and migration scripts must be manually verified.

Steps to take before running:

1. Identify the database type from the original `application.properties`.
2. Confirm the migrated Go code uses a compatible driver (e.g., `database/sql` with `lib/pq`, `go-sql-driver/mysql`, etc.).
3. Create the database and run any schema migrations manually.

```bash
# Example — adjust for your actual database:
createdb smartcontact_db
# Run schema SQL if present:
psql -d smartcontact_db -f schema.sql
```

### 5. Run the application

> **⚠️ No run command was detected during migration.** The entry point should be in `main.go` or equivalent. Verify before running.

```bash
go build -o smartcontact ./...
./smartcontact
```

Or run directly:

```bash
go run main.go
```

The server port is not confirmed — check your configuration or look for a hardcoded default in the migrated source.

---

## Running Tests

> **⚠️ No test command was detected.** The original test class (`SmartContactApplicationTests.java`, `UserServiceImpTest.java`) was migrated with low confidence. Confirm test files exist and pass before relying on them.

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

For coverage:

```bash
go test -cover ./...
```

---

## Environment Variables

No environment variables were automatically detected during migration. The table below reflects expected variables based on a typical Spring-to-Go CRUD API migration. **Manually verify each one against the migrated source.**

| Variable | Description | Required | Default |
|---|---|---|---|
| `SERVER_PORT` | Port the HTTP server listens on | No | `8080` (assumed) |
| `DB_HOST` | Database host | Yes | — |
| `DB_PORT` | Database port | Yes | — |
| `DB_NAME` | Database name | Yes | — |
| `DB_USER` | Database username | Yes | — |
| `DB_PASSWORD` | Database password | Yes | — |

> Update this table once you have reviewed `application.properties` and the migrated configuration code.

---

## Architecture Overview

The migrated Go project follows a layered structure derived from the original Spring package layout:

```
.
├── main.go                         # Application entry point (replaces SmartContactApplication.java)
├── go.mod
├── go.sum
├── controller/
│   └── user_controller.go          # HTTP handlers (migrated from UserController.java)
├── service/
│   └── user_service.go             # Business logic (migrated from UserServiceImp)
├── model/
│   └── error_message.go            # Error response model (migrated from ErrorMessage.java)
├── error/
│   └── exception_handler.go        # Global error handling middleware (migrated from RestResponseEntityExceptionHandling.java)
└── config/
    └── config.go                   # App configuration (migrated from application.properties)
```

**Key differences from Spring's structure:**

- Spring's `@RestController` / `@RequestMapping` annotations are replaced by explicit handler registration on `net/http` ServeMux.
- Spring's `@ControllerAdvice` global exception handler is replaced by a middleware wrapper or error-returning handler pattern in Go.
- Dependency injection (Spring IoC) is replaced by explicit struct initialization and manual wiring in `main.go`.
- Spring Boot auto-configuration (datasource, JPA, etc.) has no equivalent — all setup is explicit.

---

## Migration Notes

This project was migrated automatically from **Java 8+/Spring Boot** to **Go/standard library**. The overall migration confidence score is **0%**, meaning no file should be considered correct without manual review.

### What changed

| Concern | Spring (original) | Go (migrated) |
|---|---|---|
| Entry point | `@SpringBootApplication` main class | `main.go` with `http.ListenAndServe` |
| Routing | `@RequestMapping`, `@GetMapping`, etc. | `http.HandleFunc` / `ServeMux` |
| DI / IoC | `@Autowired`, `@Component`, `@Service` | Manual struct construction |
| Error handling | `@ControllerAdvice` + `ResponseEntityExceptionHandler` | Middleware or explicit error returns |
| ORM / DB | Spring Data JPA / Hibernate | `database/sql` (driver TBD) |
| Configuration | `application.properties` | Environment variables or config struct |
| Build | Maven (`pom.xml`) | Go Modules (`go.mod`) |
| Testing | JUnit + Spring Test | `testing` standard library |

### `npm install` anomaly

The migration tooling detected `npm install` as an install command. This likely indicates a misconfiguration in the migration pipeline. There is no expected Node.js dependency in a Go CRUD API. Investigate whether this was produced in error before running it.

### `pom.xml`

The original Maven build file was included in the migration scope with low confidence. Its Go equivalent (`go.mod`) must be manually verified to ensure all necessary dependencies (database drivers, etc.) are present.

---

## Known Limitations

No components were marked as entirely unmigrable. However, given the **0% confidence score across all 14 modules**, the following functional areas carry the highest risk of being incomplete or incorrect:

- **Database layer:** Spring Data JPA repositories have no direct Go equivalent. Any auto-generated CRUD queries must be manually rewritten as SQL statements.
- **Exception handling:** Spring's `@ControllerAdvice` pattern requires a deliberate Go middleware implementation — the auto-migrated version may not correctly intercept all error paths.
- **Configuration binding:** `application.properties` values (server port, datasource URL, etc.) may not be correctly mapped to Go config loading.
- **Test coverage:** JUnit tests with Spring context loading cannot be mechanically translated. The migrated tests likely do not compile or pass without significant rework.

---

## Manual Review Required

The following files were flagged as low confidence and **must be manually reviewed and corrected** before the application is run or deployed:

| File | Original Purpose | Risk |
|---|---|---|
| `pom.xml` | Maven build config → `go.mod` | Missing Go dependencies, wrong module path |
| `src/main/java/com/smartContact/SmartContactApplication.java` | Spring Boot entry point → `main.go` | Server setup, wiring, startup logic may be wrong |
| `src/main/java/com/smartContact/model/ErrorMessage.java` | Error response model | Struct fields, JSON tags may be incorrect |
| `src/main/resources/application.properties` | App config → config struct / env vars | All runtime configuration values |
| `src/test/java/com/smartContact/SmartContactApplicationTests.java` | Spring context load test | Likely does not compile; needs full rewrite |
| `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` | Global exception handler | Error middleware pattern may be non-functional |
| `src/test/java/com/smartContact/service/UserServiceImpTest.java` | Service unit tests | Test logic and assertions need full rewrite |
| `src/main/java/com/smartContact/Controller/UserController.java` | REST endpoints | Route registration, request parsing, response writing |

### Recommended review order

1. `main.go` — confirm the server starts at all
2. Config / `application.properties` migration — nothing else works without correct DB config
3. `UserController` — verify all routes are registered and handlers parse requests correctly
4. `ErrorMessage` model and exception handler — confirm error responses match the original contract
5. Service and tests last — once the above layers are confirmed correct

---

## Contributing

Since this is a migrated codebase under active review, open a PR with the specific file you have verified, what you changed, and why. Reference the original Java file in the PR description for traceability.