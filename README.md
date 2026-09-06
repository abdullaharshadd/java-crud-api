# java-crud-api

A RESTful CRUD API for managing user contacts, migrated from a Java/Spring Boot application to Go using the standard library.

## Tech Stack

- **Language:** Go
- **Framework:** Go standard library (`net/http`)
- **Database:** Relational database via `DATABASE_URL` connection string
- **Build tooling:** Go modules

## Prerequisites

- Go 1.18 or later
- A running relational database instance accessible via `DATABASE_URL`
- `git`

## Getting Started

Follow these steps in order from a fresh clone:

### 1. Install dependencies

```bash
go mod download
```

### 2. Configure environment variables

Create a `.env` file or export the variables in your shell. See the [Environment Variables](#environment-variables) section for details.

```bash
export DATABASE_URL="<your-database-connection-string>"
export PORT="8080"
```

### 3. Run the application

```bash
go run cmd/server/main.go
```

The server will start on the port defined by `PORT`.

## Running Tests

```bash
go test ./...
```

> **Note:** Ensure `DATABASE_URL` is set before running tests if any tests require a live database connection.

## Environment Variables

| Variable       | Required | Description                                                        | Example                                          |
|----------------|----------|--------------------------------------------------------------------|--------------------------------------------------|
| `DATABASE_URL` | Yes      | Full connection string for the database                            | `postgres://user:password@localhost:5432/mydb`   |
| `PORT`         | Yes      | Port on which the HTTP server listens                              | `8080`                                           |

## Architecture Overview

The migrated codebase follows a layered structure consistent with standard Go project conventions:

```
cmd/
  server/
    main.go          # Application entry point
```

The original Spring Boot layers map to Go as follows:

| Spring Layer              | Go Equivalent                          |
|---------------------------|----------------------------------------|
| `@RestController`         | HTTP handler functions (`net/http`)    |
| `@Service`                | Service structs with method receivers  |
| `@Repository` / JPA       | Data access layer (repository pattern) |
| `@ControllerAdvice`       | Centralized error handling middleware  |
| Spring `application.properties` | Environment variables             |
| Spring Boot auto-wiring   | Manual dependency injection in `main.go` |

## Migration Notes

This project was migrated from **Java 11+ / Spring Boot** to **Go / standard library**. Key changes:

- **Dependency injection:** Spring's `@Autowired` and component scanning replaced with explicit struct instantiation and manual wiring in `main.go`.
- **Routing:** Spring MVC `@RequestMapping` / `@GetMapping` etc. replaced with `net/http` `ServeMux` route registration.
- **ORM:** Spring Data JPA / Hibernate removed; database access is now handled directly without an ORM layer.
- **Error handling:** `@ControllerAdvice` / `RestResponseEntityExceptionHandler` replaced with Go middleware or handler-level error responses.
- **Configuration:** `application.properties` replaced with environment variables (`DATABASE_URL`, `PORT`).
- **Testing:** JUnit / Spring Boot Test replaced with Go's built-in `testing` package.
- **Build:** Maven (`pom.xml`) replaced with Go modules (`go.mod` / `go.sum`).

## Known Limitations

- **Overall migration confidence is 0%.** The automated migration could not produce verified, working output. The entire codebase must be treated as a starting point requiring substantial manual review and correction before it is production-ready.
- No database migration tooling (e.g., Flyway, Liquibase from the original project) has been carried over. Schema setup must be handled manually.
- The original Spring Data JPA repository behaviour (automatic query generation, pagination, etc.) has no direct equivalent and must be reimplemented explicitly in Go.

## Manual Review Required

All migrated files must be manually verified. The following files are flagged as low-confidence and are the highest priority for review:

| Original Source File | Reason for Review |
|---|---|
| `src/main/java/com/smartContact/SmartContactApplication.java` | Application bootstrap and Spring context replaced with manual Go wiring |
| `src/main/java/com/smartContact/model/User.java` | JPA entity annotations removed; struct tags and schema must be verified |
| `src/main/java/com/smartContact/repository/UserDao.java` | Spring Data JPA interface has no direct Go equivalent; queries must be hand-written |
| `src/main/java/com/smartContact/service/UserService.java` | Interface definition; ensure Go interface is correctly translated |
| `src/main/java/com/smartContact/service/UserServiceImp.java` | Service implementation logic; business rules must be manually verified |
| `src/main/java/com/smartContact/Controller/UserController.java` | REST endpoint mapping, request binding, and response serialization all changed |
| `src/main/java/com/smartContact/error/UserNotFoundException.java` | Java exception replaced with Go error type; propagation paths must be checked |
| `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` | Global exception handler has no direct Go equivalent; error handling strategy must be reimplemented |
| `src/main/resources/application.properties` | All config values must be confirmed as environment variables in the Go version |
| `src/test/java/com/smartContact/SmartContactApplicationTests.java` | Spring integration test context does not exist in Go; tests must be rewritten |
| `src/test/java/com/smartContact/service/UserServiceImpTest.java` | JUnit/Mockito-based unit tests must be rewritten using Go's `testing` package |