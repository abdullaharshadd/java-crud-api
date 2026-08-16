# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `pom.xml` → `internal/pom.xml.go` (73% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (85% confidence)
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (62% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (88% confidence)
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (60% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (58% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/apperror/restresponseentityexceptionhandling.go` (49% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (90% confidence)
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (89% confidence)
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (90% confidence)
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/user_service_impl_test.go` (36% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (88% confidence)
- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (84% confidence) ⚠️ needs review

## Components that could not be automatically migrated

These components require manual implementation. The migrated code contains
`MIGRATION_NOTE` comments at the relevant locations.

### `maven-wrapper.properties` in `.mvn/wrapper/maven-wrapper.properties`
**Reason:** This is Maven build tooling configuration, not application code. It has no logical equivalent in a target application language and is tied to the Java/Maven build ecosystem.
**Suggestion:** Do not migrate. Replace with the equivalent build system configuration for the target platform (e.g., Cargo.toml, package.json, go.mod, pyproject.toml). If the project stays on the JVM, retain this file as-is.

## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 73%

### `src/main/java/com/smartContact/error/UserNotFoundException.java`
Confidence: 78%

### `src/main/java/com/smartContact/model/ErrorMessage.java`
Confidence: 62%
Issues:
  - [warning] The Go model serializes 'status' as a numeric code (e.g. 404) while the original Java/Spring model serializes it as the enum name string (e.g. "NOT_FOUND"). This is both a JSON type change (number vs string) and a value change, which will break any consumer parsing status as a string or switching on enum names. The Target Expert concedes this is real and unresolved.

### `src/main/resources/application.properties`
Confidence: 60%
Issues:
  - [warning] The migration drops ddl-auto=update and delegates schema management to the db package's ensureSchema, which (per the comment) only does CREATE TABLE IF NOT EXISTS. This covers table creation but not the additive ALTER behavior of Hibernate's 'update' mode. Whether this narrowing is a functional regression cannot be determined from this file and depends on whether the db package has a migration story.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 58%
Issues:
  - [critical] The companion smartcontactapplication_test.go file was never actually emitted. The Target Expert concedes the gap and proposes a fix, but the fix is illustrative only and explicitly depends on unverified constructor signatures (Config, NewUserService, BuildRouterWith). As it stands, the migration still contains zero executable test logic.
  - [warning] The standalone file containing only 'const smokeTestName' is still present. The Expert agrees it should be folded into the test file and deleted, but the deletion/consolidation has not actually been performed yet.

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 49%
Issues:
  - [warning] The two-probe errors.As fix (both value and pointer forms) is a sound defensive approach, but it is proposed, not confirmed applied — and the underlying receiver semantics in usernotfoundexception.go remain unverified. The core 404-vs-500 fallthrough correctness still hinges on a file the expert admits they cannot inspect.
  - [info] The extra ValidationError->400, ErrNoRowsDeleted->500, 23505->500, and repository.ErrUserNotFound->404 mappings have no Java source equivalent and remain plan-contingent. Documented, but not yet reconciled against the actual migration plan.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 36%
Issues:
  - [critical] The renamed identifier GetUserByName is an unverified assumption; the actual exported method name and signature (including whether it takes context.Context) on the migrated UserService have not been confirmed against the codebase.
  - [critical] NewUserService(dao) only compiles if the constructor accepts an interface that *mockUserDao satisfies; if it takes the concrete *repository.UserDao this will not compile.
  - [critical] The seven mock methods were guessed at a Spring JpaRepository-style surface with no evidence the migrated repository.UserDao has those exact methods/signatures; missing methods or signature mismatches will fail compilation.

### `.mvn/wrapper/maven-wrapper.properties`
Confidence: 84%
