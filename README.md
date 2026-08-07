# java-crud-api (Go Migration)

> **Migrated project:** [`abdullaharshadd/java-crud-api`](https://github.com/abdullaharshadd/java-crud-api)
> Original stack: Java / Spring Boot → **Go / standard library**

A REST API for managing user contacts (SmartContact). Provides CRUD operations over a `users` table backed by a relational database.

---

## ⚠️ Migration Confidence Warning

**Overall migration confidence: 0%** — 14 of 14 modules were processed, but the majority of modules were flagged for low confidence or contain unmigrable components. **Do not run this in production without a thorough manual review.** See [Manual Review Required](#manual-review-required) and [Known Limitations](#known-limitations) below.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (standard library) |
| HTTP routing | `net/http` (standard library) |
| ORM / DB access | [gorm](https://gorm.io) or [sqlx](https://github.com/jmoiron/sqlx) *(choose one — see Migration Notes)* |
| Database | MySQL 8 |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Testing | `testing` (standard library) + [testify](https://github.com/stretchr/testify) |

---

## Prerequisites

- Go 1.21+
- MySQL 8.x running locally or accessible remotely
- Node.js / npm (detected in setup — verify if this is actually needed; see [Migration Notes](#migration-notes))
- A database migration tool: [golang-migrate](https://github.com/golang-migrate/migrate), Flyway, or Liquibase

Install Node dependencies (if applicable):

```bash
npm install
```

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install Go dependencies

```bash
go mod tidy
```

### 3. Configure environment variables

Copy the example env file and populate it with your values:

```bash
cp .env.example .env
```

See the [Environment Variables](#environment-variables) table for all required values.

### 4. Set up the database

The original application used `spring.jpa.hibernate.ddl-auto=update` for automatic schema management. **This has been replaced.** You must run migrations explicitly before starting the application.

If using `golang-migrate`:

```bash
migrate -path ./migrations -database "mysql://USER:PASSWORD@tcp(HOST:PORT)/DBNAME" up
```

Create the initial schema migration in `./migrations/` manually — the schema must reflect the `users` table structure from the original `User.java` model. See [Known Limitations](#known-limitations) for details.

### 5. Run the application

> **Note:** No run command was detected during migration analysis. Verify the entry point in `main.go` and update this section accordingly.

```bash
go run ./cmd/server/main.go
```

or, after building:

```bash
go build -o smartcontact ./cmd/server/main.go
./smartcontact
```

The server port defaults to `:8080` unless overridden by environment variable.

---

## Running Tests

> **Warning:** The original test suite contained non-functional tests (see [Known Limitations](#known-limitations)). Tests have been partially rewritten but require manual verification before trusting results.

```bash
go test ./...
```

To run with verbose output and coverage:

```bash
go test -v -cover ./...
```

To run a specific package:

```bash
go test -v ./internal/service/...
```

---

## Environment Variables

> No environment variables were automatically detected during migration analysis. The table below is reconstructed from the original `application.properties`. **Verify these against the migrated source code.**

| Variable | Description | Example | Required |
|---|---|---|---|
| `DB_HOST` | MySQL host | `localhost` | Yes |
| `DB_PORT` | MySQL port | `3306` | Yes |
| `DB_NAME` | Database name | `smartcontact` | Yes |
| `DB_USER` | Database username | `root` | Yes |
| `DB_PASSWORD` | Database password | `secret` | Yes |
| `SERVER_PORT` | HTTP server listen port | `8080` | No |

---

## Architecture Overview

The migrated Go project follows a layered package structure that mirrors the original Spring Boot layers:

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point (migrated from SmartContactApplication.java)
├── internal/
│   ├── model/
│   │   └── user.go              # User struct (migrated from User.java — Lombok expanded manually)
│   ├── repository/
│   │   └── user_repository.go   # CRUD methods (migrated from UserDao.java — manually implemented)
│   ├── service/
│   │   ├── user_service.go      # UserService interface
│   │   └── user_service_impl.go # UserServiceImp logic (migrated from UserServiceImp.java)
│   └── handler/
│       └── user_handler.go      # HTTP handlers (migrated from UserController.java)
├── migrations/
│   └── 0001_create_users.up.sql # Schema migration (replaces hibernate ddl-auto=update)
├── go.mod
├── go.sum
└── .env.example
```

**Request flow:**

```
HTTP Request → handler (user_handler.go)
                 → validation (go-playground/validator)
                 → service (user_service_impl.go)
                 → repository (user_repository.go)
                 → MySQL
```

Error responses (e.g., user not found → 404) are handled explicitly inside each handler function, replacing the original Spring `@ControllerAdvice`/`@ExceptionHandler` mechanism.

---

## Migration Notes

### What changed from the original Spring Boot codebase

| Area | Original (Java/Spring) | Migrated (Go) |
|---|---|---|
| Application bootstrap | `@SpringBootApplication`, `SpringApplication.run()` | `main()` in `cmd/server/main.go`, manual HTTP server setup |
| Dependency injection | Spring IoC container (`@Autowired`, `@Service`, `@Repository`) | Constructor injection via Go structs and interfaces |
| ORM / data access | Spring Data JPA + Hibernate (`JpaRepository`) | gorm or sqlx with explicit method implementations |
| Schema management | `hibernate.ddl-auto=update` (runtime auto-sync) | Explicit SQL migration files (golang-migrate recommended) |
| Request validation | JSR-380 (`@Valid`, `@NotNull`, etc. on model) | `go-playground/validator` struct tags on `User` struct |
| Exception handling | `@ControllerAdvice` + `@ExceptionHandler` | Explicit error returns and `http.Error()` calls in handlers |
| Boilerplate generation | Lombok (`@Data`, `@Getter`, `@Setter`, etc.) | All fields, getters/setters, and constructors written explicitly in Go structs |
| Query derivation | Spring Data method-name query generation (`findByName`) | Explicit SQL string: `SELECT * FROM users WHERE name = ?` |
| Testing framework | JUnit 5 + Mockito + `@SpringBootTest` | `testing` package + testify/mock or gomock, no container bootstrap |
| Build system | Maven (`pom.xml`) | Go modules (`go.mod`) |

### Note on detected `npm install`

The migration analysis detected `npm install` as a setup command. This is unexpected for a Java-to-Go migration. Verify whether the migrated project includes any frontend assets or tooling that genuinely requires Node.js. If not, this command can be ignored.

---

## Known Limitations

The following components could not be automatically migrated and require manual implementation:

### 1. JpaRepository inherited CRUD methods
**File:** `src/main/java/com/smartContact/repository/UserDao.java`
**Reason:** Spring generates these methods at runtime from an interface signature. No source code exists to translate.
**Action required:** The Go repository (`internal/repository/user_repository.go`) must implement each needed method explicitly. With gorm: use `db.Create()`, `db.First()`, `db.Find()`, `db.Save()`, `db.Delete()`. With sqlx: write explicit SQL per method.

### 2. `findByName` query derivation
**File:** `src/main/java/com/smartContact/repository/UserDao.java`
**Reason:** The SQL query was derived implicitly from the method name by Spring; no SQL exists in source.
**Action required:** Implement as explicit SQL. Example with gorm: `db.Where("name = ?", name).First(&user)`. Return `(User, error)` and handle the not-found case explicitly (check `errors.Is(err, gorm.ErrRecordNotFound)`).

### 3. `@Valid` bean validation
**File:** `src/main/java/com/smartContact/Controller/UserController.java`
**Reason:** JSR-380 annotation-driven validation has no automatic equivalent in Go.
**Action required:** Add `validate` struct tags to the `User` struct in `internal/model/user.go` and call `validator.Validate.Struct(user)` in the handler before passing to the service layer.

### 4. `UserNotFoundException` exception-to-HTTP mapping
**File:** `src/main/java/com/smartContact/Controller/UserController.java`
**Reason:** HTTP status code mapping for this exception was defined in a separate `@ControllerAdvice` not present in the source files provided.
**Action required:** Locate the original exception handler (if available), then add explicit error-to-status mapping in `user_handler.go`. A not-found error should write `http.StatusNotFound` (404).

### 5. `@SpringBootTest` smoke test
**File:** `src/test/java/com/smartContact/SmartContactApplicationTests.java`
**Reason:** Relies on Spring's `ApplicationContext` bootstrapping, which has no Go equivalent.
**Action required:** Replace with a native Go smoke test that initializes the application and asserts it starts without error (e.g., test that the HTTP server responds to a health-check endpoint).

### 6. `UserServiceImpTest` — Mockito on real bean
**File:** `src/test/java/com/smartContact/service/UserServiceImpTest.java`
**Reason:** The original test applies `Mockito.when()` to a real Spring-managed bean rather than a mock. This is a bug in the original test — the stubbing has no effect. There is no correct behavior to translate.
**Action required:** Rewrite from scratch. Define a `UserRepository` interface in Go, create a mock using `testify/mock` or `gomock`, stub the relevant methods, inject the mock into `UserServiceImpl` via constructor, and assert service behavior per test case.

### 7. `@BeforeAll` lifecycle in tests
**File:** `src/test/java/com/smartContact/service/UserServiceImpTest.java`
**Reason:** JUnit 5 `@BeforeAll` semantics (particularly `PER_CLASS` lifecycle behavior) have no Go equivalent. The original usage is also non-functional due to the real-bean issue above.
**Action required:** Perform all test setup inside individual test functions for isolation. Use `TestMain` in the package only if package-level setup is genuinely needed.

### 8. `spring.jpa.hibernate.ddl-auto=update`
**File:** `src/main/resources/application.properties`
**Reason:** Hibernate-specific runtime schema synchronization.
**Action required:** Create explicit migration files in `./migrations/`. Do not rely on any ORM's auto-migrate feature in production.

### 9. Lombok
**File:** `pom.xml` and all model/entity classes
**Reason:** Lombok is a Java-only compile-time annotation processor.
**Action required:** All Lombok-generated members (`@Data`, `@Getter`, `@Setter`, `@NoArgsConstructor`, `@AllArgsConstructor`, `@Builder`) must be expanded explicitly in the Go `User` struct. Verify `internal/model/user.go` has all required fields and that no generated behavior was silently dropped.

---

## Manual Review Required

The following files were flagged as low-confidence migrations. A developer must read each file carefully and verify correctness before the application is considered functional:

| File | Primary concern |
|---|---|
| `internal/model/user.go` | Lombok expansion — verify all fields, types, and zero values are correct |
| `internal/service/user_service.go` | Interface definition matches all methods actually used by handlers |
| `internal/service/user_service_impl.go` | Business logic translated correctly; error handling is explicit and correct |
| `internal/handler/user_handler.go` | Routing, JSON marshaling, validation, and error-to-status mapping are all correct |
| `cmd/server/main.go` | Application wiring (DB init, router setup, dependency injection chain) is complete |
| `internal/service/user_service_impl_test.go` | Test is a rewrite, not a translation — assert it actually tests meaningful behavior |
| `app_test.go` (SmartContactApplicationTests) | Smoke test must validate actual startup, not just compile |
| `config/` or env loading | `application.properties` values must be loaded from environment variables or config file |
| `go.mod` | Verify all dependencies are correct and no Java-specific artifacts remain |

---

## Contributing

When addressing the manual review items above, open one PR per component for easier review. Reference the specific limitation or flagged file in the PR description.

---

## License

See original repository: [abdullaharshadd/java-crud-api](https://github.com/abdullaharshadd/java-crud-api)