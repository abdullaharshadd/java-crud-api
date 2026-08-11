```markdown
# smart-contact-api (Go)

A REST API for managing user/contact records. Migrated from the original
[abdullaharshadd/java-crud-api](https://github.com/abdullaharshadd/java-crud-api)
Spring Boot application to idiomatic Go using the standard library.

> **⚠ Migration confidence: 0% overall — every file in this repository
> requires manual review before this code can be considered production-ready.**
> See [Manual Review Required](#manual-review-required) and
> [Known Limitations](#known-limitations) below.

---

## Tech Stack

| Layer | Original (Java) | Migrated (Go) |
|---|---|---|
| Language | Java 17 | Go 1.21+ |
| Framework | Spring Boot | `net/http` (standard library) |
| ORM / DB layer | Spring Data JPA + Hibernate | To be chosen (see notes) |
| Database | MySQL | MySQL |
| Validation | Bean Validation (JSR-380) | Manual / `go-playground/validator` |
| Error handling | `@ControllerAdvice` | Middleware / handler wrappers |
| Build | Maven | Go modules (`go.mod`) |
| Tests | JUnit 5 + Mockito + `@SpringBootTest` | `testing` + `testify/mock` or `gomock` |

---

## Prerequisites

- Go 1.21 or later — <https://go.dev/dl/>
- MySQL 8.x running and accessible
- `npm` — only required if the detected `npm install` setup step applies to
  an auxiliary tooling script in this repo (verify before running)
- `git`

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/java-crud-api.git
cd java-crud-api
```

### 2. Install dependencies

> **Note:** `npm install` was detected as an install command by the migration
> tool. Verify whether a `package.json` exists in the repo root and what it
> installs. If it is unrelated to the Go application, skip this step.

```bash
npm install        # only if applicable — verify first
go mod download    # download Go module dependencies
```

### 3. Configure environment

No environment variables were detected automatically. The original application
used `src/main/resources/application.properties` for database connection
settings. You must map those values to the Go configuration mechanism chosen
for this project.

Copy the example config and edit it:

```bash
cp .env.example .env   # create this file if it does not exist
```

At minimum you will need MySQL connection details — see
[Environment Variables](#environment-variables) below.

### 4. Database setup

The original project relied on `spring.jpa.hibernate.ddl-auto=update` to
manage the schema automatically. **This behavior has not been replicated.**
You must set up the schema manually:

1. Create the database:

```sql
CREATE DATABASE smart_contact;
```

2. Apply the schema. Because no migration tool was configured during
   migration, you must either:
   - Write an initial SQL migration file from the original entity classes, or
   - Integrate a migration tool such as
     [golang-migrate](https://github.com/golang-migrate/migrate) and create
     the first migration.

### 5. Run the application

```bash
go run ./cmd/server        # adjust path to match actual entry point
```

> **⚠ The entry point path (`./cmd/server`) must be verified.** The migration
> did not detect a confirmed `main.go` location. Check the repository
> structure and update this command accordingly.

---

## Running Tests

```bash
go test ./...
```

To run with verbose output and race detection:

```bash
go test -v -race ./...
```

> **⚠ Tests require significant manual rewriting.** See
> [Known Limitations](#known-limitations) for details on why the original
> test classes could not be cleanly migrated.

---

## Environment Variables

No variables were confirmed by the automated migration. The table below is
derived from the original `application.properties`. **Verify and extend this
table after reviewing the migrated configuration code.**

| Variable | Description | Example | Required |
|---|---|---|---|
| `DB_HOST` | MySQL host | `localhost` | Yes |
| `DB_PORT` | MySQL port | `3306` | Yes |
| `DB_NAME` | Database name | `smart_contact` | Yes |
| `DB_USER` | Database username | `root` | Yes |
| `DB_PASSWORD` | Database password | `secret` | Yes |
| `SERVER_PORT` | Port the HTTP server listens on | `8080` | No |

---

## Architecture Overview

The migrated project follows a layered structure that mirrors the original
Spring package layout, translated to Go conventions:

```
.
├── cmd/
│   └── server/
│       └── main.go          # application entry point
├── internal/
│   ├── handler/             # HTTP handlers (replaces UserController)
│   │   └── user_handler.go
│   ├── service/             # business logic (replaces UserService / UserServiceImp)
│   │   └── user_service.go
│   ├── repository/          # data access (replaces UserDao)
│   │   └── user_repository.go
│   ├── model/               # domain structs (replaces User entity + ErrorMessage)
│   │   ├── user.go
│   │   └── error_message.go
│   └── middleware/          # error handling (replaces RestResponseEntityExceptionHandling)
│       └── error_handler.go
├── go.mod
├── go.sum
└── README.md
```

> **⚠ This layout is the intended target structure.** The actual files
> generated by the migration tool may differ. Reconcile the generated output
> against this layout before running the application.

**Request flow:**

```
HTTP Request
  └─► middleware (error handling, logging)
        └─► handler (parse request, call service, write response)
              └─► service (business logic, validation)
                    └─► repository (SQL queries)
                          └─► MySQL
```

---

## Migration Notes

Summary of what changed when moving from Spring Boot to Go.

### Build system

- `pom.xml` and `.mvn/wrapper/maven-wrapper.properties` have no Go
  equivalent. The project now uses `go.mod` / `go.sum`. Dependencies must be
  chosen and added manually — no Spring Boot starters exist in the Go
  ecosystem.

### Dependency injection

- Spring's `@Autowired`, `@Service`, `@Repository`, and `@RestController`
  annotations are gone. Dependencies are wired explicitly via constructor
  functions or struct initialization in `main.go`.

### Lombok

- `@Data`, `@Getter`, `@Setter`, `@NoArgsConstructor`, etc. have been replaced
  with plain Go structs with exported fields. No code generation is involved.

### Error handling

- `@ControllerAdvice` + `ResponseEntityExceptionHandler` have been replaced
  with HTTP middleware. However, the inherited default exception mappings from
  `ResponseEntityExceptionHandler` (400 for validation errors, 405 for method
  not allowed, 415 for unsupported media type, etc.) **must be reimplemented
  manually** — they were implicit in the original and are not present in the
  generated code.

### Schema management

- `spring.jpa.hibernate.ddl-auto=update` has been removed. Schema changes are
  now a manual responsibility. Introduce a migration tool
  (e.g., `golang-migrate`) before deploying to any persistent environment.

### Validation

- Bean Validation annotations (`@NotNull`, `@Size`, etc.) have no runtime
  equivalent in Go. Validation logic must be added explicitly, either in the
  service layer or using `github.com/go-playground/validator/v10`.

### Testing

- `@SpringBootTest` context loading does not exist in Go. Tests construct
  dependencies directly. Mockito stubs have been replaced with interface-based
  mocking via `testify/mock` or `gomock`.

---

## Known Limitations

The following components could not be fully or correctly migrated and require
manual intervention.

| File | Component | Reason |
|---|---|---|
| `.mvn/wrapper/maven-wrapper.properties` | maven-wrapper.properties | Maven-specific build tooling; no Go equivalent. File is not present in the migrated project. |
| `pom.xml` | Entire file | Maven POM is JVM-only. Dependencies must be manually replaced with Go equivalents (Gin or `net/http`, GORM or `database/sql`, MySQL driver, validator). |
| `pom.xml` | `spring-boot-maven-plugin` | Executable JAR packaging; replaced by `go build`. |
| `pom.xml` | `lombok` | Compile-time annotation processor; dropped. Use plain Go structs. |
| `src/main/resources/application.properties` | `spring.jpa.hibernate.ddl-auto=update` | Hibernate auto-DDL has no equivalent. Schema migrations must be managed explicitly. |
| `src/main/java/.../RestResponseEntityExceptionHandling.java` | `ResponseEntityExceptionHandler` superclass | Implicit default exception mappings (validation errors, unsupported media type, etc.) are not visible in the source and were not migrated. These must be reimplemented as explicit middleware cases. |
| `src/test/java/.../UserServiceImpTest.java` | `setUp` Mockito stubbing | Mockito stubbing on a `@Autowired` real bean is a Spring/Mockito-specific pattern and does not translate. Tests must be rewritten to inject a mock repository into the service struct. |
| `src/test/java/.../UserServiceImpTest.java` | `@SpringBootTest` | Boots the full Spring context; has no Go equivalent. Replace with direct struct construction. |

---

## Manual Review Required

**Every migrated file requires review** (overall migration confidence is 0%).
The files below have been specifically flagged as high-risk.

| File | What to verify |
|---|---|
| `.mvn/wrapper/maven-wrapper.properties` | Confirm it has been removed from the Go project. No translation is possible or needed. |
| `pom.xml` | Confirm all required dependencies have been identified and added to `go.mod` with idiomatic Go equivalents. |
| `src/main/java/com/smartContact/model/ErrorMessage.java` | Verify the migrated `error_message.go` struct matches the JSON field names and types expected by API consumers. |
| `src/main/resources/application.properties` | Confirm all config values (DB URL, pool settings, server port) are surfaced in the Go config layer and documented in the environment variables table. |
| `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` | Manually enumerate all exception types handled (including inherited ones) and implement corresponding cases in the Go error middleware. |
| `src/main/java/com/smartContact/repository/UserDao.java` | Verify all query methods are correctly translated to SQL with the chosen Go database library. Pay attention to derived query method names (e.g., `findByEmail`). |
| `src/main/java/com/smartContact/service/UserService.java` | Confirm the service interface is complete and matches all methods used by the controller. |
| `src/main/java/com/smartContact/service/UserServiceImp.java` | Verify business logic, transaction boundaries (originally managed by Spring's `@Transactional`), and error propagation. |
| `src/main/java/com/smartContact/Controller/UserController.java` | Verify all routes, HTTP methods, path variables, request body binding, and response status codes match the original API contract. |
| `src/test/java/com/smartContact/SmartContactApplicationTests.java` | Replace Spring context smoke test with a Go equivalent (e.g., confirm the server starts and returns 200 on a health endpoint). |
| `src/test/java/com/smartContact/service/UserServiceImpTest.java` | Rewrite entirely: construct the service with a mock repository, stub repository methods, and assert service behavior without any framework context. |

---

## Contributing

1. Resolve all items in [Manual Review Required](#manual-review-required)
   before merging any feature work.
2. Ensure `go vet ./...` and `go test -race ./...` pass with no errors.
3. Add a SQL migration file under `migrations/` for any schema change.
```