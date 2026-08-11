# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (92% confidence)
- `pom.xml` → `internal/pom.xml.go` (70% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (43% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (90% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (90% confidence)
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (89% confidence)
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (66% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (61% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/apperror/restresponseentityexceptionhandling.go` (88% confidence)
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (70% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (76% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (55% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/service_test/userserviceimptest.go` (43% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (86% confidence)
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 70%
Issues:
  - [info] The migration switches the persistence driver from MySQL to PostgreSQL (lib/pq, $1/$2 placeholders, RETURNING id), which conflicts with the stated project invariant requiring MySQL at runtime. The pom.xml manifest documents this divergence honestly but does not resolve the underlying conflict.

### `src/main/java/com/smartContact/SmartContactApplication.java`
Confidence: 43%
Issues:
  - [critical] The submitted code still uses degraded-mode 503 handlers instead of aborting on DB/schema failure. The Target Expert concedes this and proposes the correct fix, but the actual pasted code is still the pre-fix version.
  - [critical] The submitted buildRouter returns only http.Handler with no path for main to close db, causing a genuine connection leak. Target Expert concedes this and provides correct cleanup-closure plus signal-based shutdown wiring, but the actual code remains unfixed.

### `src/main/resources/application.properties`
Confidence: 66%
Issues:
  - [info] The relocation of ddl-auto=update behavior to model.EnsureUserSchema is documented but unverifiable from this file; the actual existence and startup invocation of that call must be confirmed in smartcontactapplication.go.
  - [warning] The dialect switch is correctly flagged and the DSN is a valid PostgreSQL string, but repository-layer SQL correctness ($1 placeholders, RETURNING vs LAST_INSERT_ID(), SERIAL vs AUTO_INCREMENT) remains unverified and is external to this file.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 61%
Issues:
  - [critical] The Target Expert correctly diagnoses the missing blank import causing sql.Open to fail with 'unknown driver' and the test to skip unconditionally, but the response only proposes the fix — it does not confirm the driver import was actually added to the migrated file. Until the blank import is committed (matching the production bootstrap's driver/DSN scheme), the smoke test remains silently defeated.

### `src/main/java/com/smartContact/repository/UserDao.java`
Confidence: 70%
Issues:
  - [info] The Target Expert acknowledges the behavioral drift (Java errors on multiple matches via IncorrectResultSizeDataAccessException, Go silently returns the first row) but conditions the fix on whether 'name' has a UNIQUE constraint, which is not confirmed. As-is, the migrated code still silently returns the first match rather than failing loud.

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 76%

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 55%
Issues:
  - [info] The provided migrated file is only a documentation/mapping comment file with no actual implementation. The real logic is claimed to reside in userservice.go, which was not provided for review.
  - [info] Cannot verify that a missing row is correctly wrapped as UserNotFoundError since the implementation file is not included here.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 43%
Issues:
  - [critical] Test fabricates API surface (svc.FetchUserById, apperr.UserNotFoundError) not present in the original source; still present in file as a compile/contract risk until actually removed.
  - [critical] Test references unverified svc.DeleteUser and repository.ErrUserNotFound symbols; speculative contract, compile risk.
  - [warning] Asserts an unverified (nil, nil) not-found contract admitted to be a guess; will pass by luck or fail against correct behaviour.
