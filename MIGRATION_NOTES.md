# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (97% confidence)
- `pom.xml` → `internal/pom.xml.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (76% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (94% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (58% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (70% confidence) ⚠️ needs review
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (78% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests_test.go` (66% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling/restresponseentityexceptionhandling.go` (88% confidence)
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (70% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (73% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (44% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/service_test.go` (73% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (82% confidence) ⚠️ needs review
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 78%

### `src/main/java/com/smartContact/SmartContactApplication.java`
Confidence: 76%
Issues:
  - [warning] The fix removes the fabricated layered architecture but introduces a new invention: an HTTP server (http.NewServeMux + ListenAndServe on :8080). The 13-line source contains only SpringApplication.run(...); the port, mux, and server startup are not evidenced in the source file and are the Expert's own inference about @SpringBootApplication's runtime behavior.

### `src/main/java/com/smartContact/model/ErrorMessage.java`
Confidence: 58%
Issues:
  - [critical] The Target Expert concedes the point and lays out correct fixes (Option A custom marshaler emitting the Spring enum name, or Option B documented breaking change), but neither has been applied to the shipped code. The migration still serializes status as a numeric int (404) instead of the Java enum name string ("NOT_FOUND").

### `src/main/java/com/smartContact/model/User.java`
Confidence: 70%
Issues:
  - [info] The Target Expert conceded and provided a corrected comment, but explicitly stated the fix is not yet applied ('previously conceded this but never applied the fix'). The proposed rewrite is correct, but it remains a proposal in the response rather than confirmed-applied code.
  - [warning] The Target Expert proposes Option A (return standard error) which resolves the compilation concern, but this too is presented as a proposed change rather than confirmed-applied. Changing the return type from *ErrorMessage to error is a signature change that any callers of Validate() must be updated to match; the response does not verify caller impact.

### `src/main/resources/application.properties`
Confidence: 78%

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 66%
Issues:
  - [warning] The test assumes a function `buildRouter()` with a nilable return type exists, but this is unverified against the real source. If the name or signature differs (e.g. returns (router, error) or a value type), the test will fail to compile. This is a real, unresolved compilation risk that the expert acknowledges but has not resolved.

### `src/main/java/com/smartContact/repository/UserDao.java`
Confidence: 70%
Issues:
  - [info] Omitted methods (count, delete-by-entity, existsById, etc.) remain unimplemented pending a service-layer usage audit that has not yet been performed. Whether this is a real defect depends on actual caller usage, which is still unverified.

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 73%
Issues:
  - [warning] The doc comment promises an empty slice is returned when there are no users, but the implementation passes through whatever s.repo.FindAll returns. If FindAll returns nil, callers get nil, contradicting the documented contract.
  - [warning] The Java void deleteUser(id) was a silent no-op for a nonexistent id, but the migration adds a FindByID guard that returns ErrUserNotFound. This is an undocumented observable behavior change not reflected in any MIGRATION_NOTE.
  - [info] The condition `apperror.IsUserNotFound(err) || errors.Is(err, apperror.ErrUserNotFound)` is redundant if IsUserNotFound already wraps errors.Is. Dead code, not a correctness bug.

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 44%
Issues:
  - [info] The Java method returned the repository result verbatim, returning null when no user is found (no error). The migrated GetUserByName returns an ErrUserNotFound sentinel instead of a nil/absent value. This is a deliberate, documented paradigm change and is acceptable for absence-signalling, but it does alter the observable contract: Java callers treated null as a valid non-error outcome.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 73%

### `src/main/java/com/smartContact/Controller/UserController.java`
Confidence: 82%
