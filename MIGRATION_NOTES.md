# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (96% confidence)
- `pom.xml` → `internal/pom.xml.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (86% confidence)
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (94% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (91% confidence)
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (94% confidence)
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (78% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (80% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling/restresponseentityexceptionhandling.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (72% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (84% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (83% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/userserviceimptest.go` (88% confidence)
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (89% confidence)

## Components that could not be automatically migrated

These components require manual implementation. The migrated code contains
`MIGRATION_NOTE` comments at the relevant locations.

### `maven-wrapper.properties` in `.mvn/wrapper/maven-wrapper.properties`
**Reason:** This file is specific to the Maven/JVM build system and has no logical equivalent to migrate; it is pure build-tool configuration.
**Suggestion:** If staying on the JVM with Maven, keep as-is. If migrating to another language/ecosystem, discard this file and configure the target's native build tool and version-pinning mechanism instead.

### `lombok` in `pom.xml`
**Reason:** Lombok is a JVM compile-time annotation processor that generates Java bytecode; it has no meaning outside the Java/Kotlin ecosystem.
**Suggestion:** Do not migrate the dependency. When migrating annotated model classes, explicitly generate the boilerplate (getters/setters/equals/hashCode/builders/constructors) using the target language's idioms (e.g., Python dataclasses/attrs, TypeScript class properties, Go struct fields, Kotlin data classes).

### `spring-boot-maven-plugin` in `pom.xml`
**Reason:** Maven-specific build/packaging plugin producing an executable JVM JAR; no cross-language equivalent.
**Suggestion:** Replace with the target ecosystem's build/bundling tool (npm/webpack/esbuild, PyInstaller/poetry build, go build, cargo build) and configure the runnable artifact there.

### `spring-boot-starter-parent (BOM version management)` in `pom.xml`
**Reason:** Maven parent-POM dependency version resolution is Maven-specific and cannot be auto-translated.
**Suggestion:** Manually pin dependency versions in the target's manifest file and reintroduce framework-equivalent packages explicitly.

### `spring.jpa.hibernate.ddl-auto=update` in `src/main/resources/application.properties`
**Reason:** Automatic schema generation/update from entity annotations is a Hibernate-specific runtime behavior with no direct equivalent in most non-JVM ORMs.
**Suggestion:** Replace with an explicit database migration tool in the target stack (Flyway/Liquibase for JVM, Alembic for Python, Prisma/TypeORM migrations for Node, EF Core migrations for .NET). Generate initial schema from existing DB.

### `SmartContactApplicationTests` in `src/test/java/com/smartContact/SmartContactApplicationTests.java`
**Reason:** It is Spring-specific framework glue (@SpringBootTest) with no portable logic; its sole purpose is validating Spring's context bootstrap, which has no meaning outside Spring.
**Suggestion:** Do not port directly. Recreate a minimal 'application startup / health check' test using the target framework's testing conventions, or omit if the target has its own bootstrap validation.

### `setUp / Mockito.when(userServiceImp.getUserNameByName(...))` in `src/test/java/com/smartContact/service/UserServiceImpTest.java`
**Reason:** Mockito is being misused against a real Spring-managed bean (@Autowired, not @MockBean), which is invalid and would fail at runtime. There is no direct equivalent, and copying it would produce a non-functional test.
**Suggestion:** Manually rewrite: define a UserDao/repository interface in Go, provide a fake/mock implementation returning the test User, inject it via constructor into the service, then assert the returned user's name. Use testify/assert and (optionally) testify/mock or a hand-rolled stub.

### `@BeforeAll non-static method` in `src/test/java/com/smartContact/service/UserServiceImpTest.java`
**Reason:** Requires @TestInstance(Lifecycle.PER_CLASS) which is absent, making the setup lifecycle invalid; this framework-specific magic has no Go equivalent.
**Suggestion:** Use TestMain for one-time setup or, preferably, per-test setup inside each Go test function or a shared helper that builds the fixture.

## Observer agent findings

The Observer agent monitored the migration and identified these patterns:

- **After 3 modules:** Could not parse observer output
- **After 6 modules:** Could not parse observer output
- **After 9 modules:** Could not parse observer output
- **After 12 modules:** Could not parse observer output

## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 78%

### `src/main/resources/application.properties`
Confidence: 78%
Issues:
  - [info] The original relies on Hibernate to auto-create/update schema at startup. The Go migration only records the DDLAuto value as a string and performs no actual schema management. This is an expected and correctly-documented migration gap (no Go ORM auto-migration), not a code defect.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 80%

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 78%

### `src/main/java/com/smartContact/repository/UserDao.java`
Confidence: 72%
Issues:
  - [info] The Java derived finder returns null on no match; the Go version returns a wrapped ErrUserNotFound error instead. This is an explicitly documented and recommended migration convention (Go idiom favors errors.Is over null checks), so it is an acceptable paradigm difference. Callers must be adapted, but behavior is preserved semantically.
  - [warning] The Java single-object derived query throws a NonUniqueResultException when multiple rows match. The Go version uses QueryRowContext which silently returns only the first row and ignores extras, so no error is raised for duplicate names.

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 84%
Issues:
  - [info] The original spec states lookup-by-name returns null/no user when no match exists (not an error). The migration returns a UserNotFoundError when the user is nil or the DAO reports not-found. This is a behavioral difference for the not-found case, but callers treating error-as-absent generally still work.

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 83%
Issues:
  - [warning] The Java findByName returns null (no error) when no user matches, and the caller must handle null. The Go migration returns a UserNotFoundError on the not-found case, changing the observable contract: a missing name is now an error rather than a null/empty result. This is a real behavioral difference but only matters if a caller relied on the null-return path.
