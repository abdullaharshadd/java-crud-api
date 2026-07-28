# java-crud-api (Go Migration)

> **⚠️ Migration Warning:** This project was automatically migrated from Java/Spring to Go/standard library with **0% overall confidence**. All migrated files require thorough manual review before use in any environment.

---

## Description

A CRUD REST API for managing smart contacts (users). The original application exposed endpoints to create, read, update, and delete user records backed by a relational database. This repository contains the Go port of that Spring Boot service.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (standard library) |
| HTTP Router | `net/http` (standard library) |
| Database Driver | To be confirmed during manual review |
| Testing | `testing` (standard library) |
| Build Tool | Go modules (`go.mod`) |

> The original project used Java 8+, Spring Boot, Spring Data JPA, and Maven (`pom.xml`).

---

## Prerequisites

- Go 1.21 or later
- A running relational database instance (PostgreSQL or MySQL — confirm from original `application.properties`)
- Node.js / npm (detected in setup plan — verify whether this is actually required or a migration artifact)
- Git

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install dependencies

```bash
npm install
```

> **⚠️ Review required:** `npm install` was detected in the setup plan but this is a Go project. Verify whether a `package.json` is intentionally present (e.g., for tooling or docs generation) or whether the correct command should be:
>
> ```bash
> go mod download
> ```

### 3. Configure environment variables

No environment variables were detected automatically. However, the original `application.properties` contained database connection details that must be wired into the Go application. Create a `.env` file or export variables directly — see the [Environment Variables](#environment-variables) section.

### 4. Database setup

No database setup command was detected during migration. Before running the application:

1. Create your database manually.
2. Apply any schema migrations. Check for SQL migration files in the repository — if none exist, reconstruct the schema from the original `User.java` model.
3. Confirm the `User` table structure matches what the Go data layer expects.

### 5. Run the application

No run command was detected during migration. Use:

```bash
go run ./...
```

or build a binary first:

```bash
go build -o smart-contact ./...
./smart-contact
```

---

## Running Tests

No test command was detected during migration. Use the standard Go test runner:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

> **⚠️ Review required:** The original project contained `SmartContactApplicationTests.java` (Spring integration test) and `UserServiceImpTest.java` (unit test). Verify that equivalent Go tests exist and cover the same cases before relying on test results.

---

## Environment Variables

No environment variables were automatically detected. Based on the original Spring `application.properties`, the following are likely required. Confirm and update this table after manual review.

| Variable | Description | Example | Required |
|---|---|---|---|
| `DB_HOST` | Database host | `localhost` | Likely yes |
| `DB_PORT` | Database port | `5432` / `3306` | Likely yes |
| `DB_NAME` | Database name | `smart_contact` | Likely yes |
| `DB_USER` | Database username | `root` | Likely yes |
| `DB_PASSWORD` | Database password | `secret` | Likely yes |
| `SERVER_PORT` | Port the HTTP server listens on | `8080` | Likely yes |

> These variables are inferred from the Java source. The Go application may use different names or read from a config file. Verify in the migrated source before deploying.

---

## Architecture Overview

The migrated Go project follows a layered structure that mirrors the original Spring package layout:

```
.
├── go.mod
├── main.go                        # Entry point (migrated from SmartContactApplication.java)
├── model/
│   ├── user.go                    # User struct (migrated from User.java)
│   └── error_message.go           # Error response struct (migrated from ErrorMessage.java)
├── repository/
│   └── user_dao.go                # Data access layer (migrated from UserDao.java)
├── service/
│   ├── user_service.go            # Service interface (migrated from UserService.java)
│   └── user_service_imp.go        # Service implementation (migrated from UserServiceImp.java)
├── controller/
│   └── user_controller.go         # HTTP handlers (migrated from UserController.java)
└── test/
    ├── main_test.go               # App-level tests (migrated from SmartContactApplicationTests.java)
    └── service/
        └── user_service_imp_test.go  # Service unit tests (migrated from UserServiceImpTest.java)
```

> **Note:** Actual file paths may differ. The above reflects the expected structure based on the original Java package layout. Verify against the actual repository contents.

The application follows a standard request lifecycle:

```
HTTP Request → controller (net/http handler) → service (business logic) → repository (DB access) → Response
```

---

## Migration Notes

This project was automatically migrated from **Java/Spring Boot** to **Go standard library**. The following changes were made:

### Framework & Dependency Injection
- Spring's `@Autowired` dependency injection was removed. Dependencies are wired manually in `main.go` or via constructor functions.
- `@RestController`, `@RequestMapping`, `@GetMapping`, `@PostMapping`, etc. annotations were replaced with `net/http` handler functions and a manual router.

### Build System
- Maven (`pom.xml`) was replaced with Go modules (`go.mod`). All Spring Boot starters and JPA dependencies were removed.

### Data Layer
- Spring Data JPA (`UserDao` extending `JpaRepository`) was replaced with manual SQL queries using Go's `database/sql` package. ORM behaviour is **not** replicated automatically — query logic must be verified.

### Model / Entity
- JPA annotations (`@Entity`, `@Id`, `@GeneratedValue`, `@Column`) were removed. The `User` struct uses plain Go fields. JSON serialisation uses `encoding/json` struct tags.

### Error Handling
- Spring's `ResponseEntity` and `@ExceptionHandler` patterns were replaced with manual HTTP status writes using `http.ResponseWriter`.

### Configuration
- `application.properties` was replaced with environment variables or a config struct. Confirm the configuration strategy in `main.go`.

### Testing
- JUnit and Spring Boot Test (`@SpringBootTest`, `@MockBean`) were replaced with Go's `testing` package. Mockito-based mocks must be replaced with Go interface mocks or a library such as `testify/mock`.

---

## Known Limitations

The following limitations are a direct result of the low-confidence (0%) automated migration:

- **No ORM:** Go standard library has no JPA equivalent. Complex queries derived from Spring Data method names (e.g., `findByEmail`) may not have been correctly translated to SQL.
- **No dependency injection framework:** Manual wiring may be incomplete or incorrect if the original Spring context had non-trivial bean scoping.
- **No validation framework:** Spring's `@Valid` / `javax.validation` constraints are not present in Go standard library. Input validation logic may be missing entirely.
- **No migration scripts:** Flyway or Liquibase (if used in the original project) have no Go equivalent in this migration. Database schema must be managed manually.
- **Transaction management:** Spring's `@Transactional` annotation has no direct equivalent. Transaction handling in the repository layer must be manually verified and implemented.
- **Test coverage unknown:** The automated migration of test files cannot be verified to be correct given 0% confidence. Do not rely on test results until tests are manually reviewed.

---

## Manual Review Required

**All 14 migrated modules require review.** The migration confidence score is 0%, meaning no file should be considered production-ready without human verification.

Priority order for review:

| Priority | File | Reason |
|---|---|---|
| 🔴 Critical | `main.go` (from `SmartContactApplication.java`) | Entry point, server setup, dependency wiring |
| 🔴 Critical | `repository/user_dao.go` (from `UserDao.java`) | All database queries must be manually written and verified |
| 🔴 Critical | `service/user_service_imp.go` (from `UserServiceImp.java`) | Core business logic; `@Transactional` behaviour lost |
| 🔴 Critical | `go.mod` (from `pom.xml`) | Dependency correctness; confirm all required packages are present |
| 🔴 Critical | Config / env setup (from `application.properties`) | DB connection and server config must work correctly |
| 🟠 High | `controller/user_controller.go` (from `UserController.java`) | HTTP routing, request parsing, response serialisation |
| 🟠 High | `model/user.go` (from `User.java`) | Field types, JSON tags, and DB column mapping |
| 🟠 High | `model/error_message.go` (from `ErrorMessage.java`) | Error response contract used across handlers |
| 🟠 High | `service/user_service.go` (from `UserService.java`) | Interface definition must match implementation |
| 🟡 Medium | `test/service/user_service_imp_test.go` (from `UserServiceImpTest.java`) | Unit tests must use Go idioms; mocks must be re-implemented |
| 🟡 Medium | `test/main_test.go` (from `SmartContactApplicationTests.java`) | Integration test approach differs significantly from Spring |

---

## Original Project

Source repository: [abdullaharshadd/java-crud-api](https://github.com/abdullaharshadd/java-crud-api)
Original stack: Java · Spring Boot · Spring Data JPA · Maven