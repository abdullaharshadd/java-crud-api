# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (97% confidence)
- `pom.xml` → `internal/pom.xml.go` (61% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (90% confidence)
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (95% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (89% confidence)
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (78% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (53% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling/restresponseentityexceptionhandling.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (86% confidence)
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (83% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (69% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/userserviceimptest.go` (76% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (88% confidence)
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 61%
Issues:
  - [warning] The migration documentation states PostgreSQL (github.com/jackc/pgx/v5) is the TARGET database and intentionally does NOT mirror the source MySQL driver. The source explicitly uses MySQL (com.mysql:mysql-connector-j) and the migration notes/global invariants state 'MySQL as the runtime database driver'. Switching the database engine is a real behavioral divergence that can affect SQL dialect, connection config, and data-type semantics.

### `src/main/java/com/smartContact/model/ErrorMessage.java`
Confidence: 78%

### `src/main/resources/application.properties`
Confidence: 78%
Issues:
  - [info] The credential defaults were silently changed from root/root (the original observable values) to postgres/postgres without a dedicated inline migration note. If the target PostgreSQL instance is provisioned with the original root credentials, the defaults would fail to connect out-of-the-box.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 53%
Issues:
  - [warning] The Target Expert provided a plausible, executable test skeleton, but it depends on constructs (NewApplication, app.Close(), app.UserDao) whose existence in the actual Go codebase is unverified. The expert explicitly notes 'If the codebase has no NewApplication constructor... that must be identified and used' — meaning the code as written may not compile. The concrete fix is not confirmed to be committed to the actual file.

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 78%
Issues:
  - [info] Correctness depends on model.NewErrorMessageFromError extracting err.Error() as the message field, which is unverified in this scope.

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 83%

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 69%
Issues:
  - [info] The compile-time assertion `var _ UserServicer = (*UserService)(nil)` and DTO types (`model.CreateUserRequest`, `model.UserResponse`) depend on `userservice.go` and the `model` package existing with matching signatures. This cannot be confirmed from this file alone, and the expert himself flags this as an unverified precondition.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 76%
