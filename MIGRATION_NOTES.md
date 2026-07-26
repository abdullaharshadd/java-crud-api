# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (82% confidence) ⚠️ needs review
- `pom.xml` → `internal/pom.xml.go` (66% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (88% confidence)
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (94% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (84% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (83% confidence) ⚠️ needs review
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (78% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (64% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling/restresponseentityexceptionhandling.go` (85% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (52% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (62% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/userserviceimptest.go` (83% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (82% confidence) ⚠️ needs review
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `.mvn/wrapper/maven-wrapper.properties`
Confidence: 82%

### `pom.xml`
Confidence: 66%
Issues:
  - [warning] The Target Expert correctly diagnosed the fabricated MySQL→PostgreSQL directive and provided an accurate MySQL-dialect replacement, but the response only quotes the proposed corrected text—it does not demonstrate that the fix was actually applied to the file. The defect remains open until the corrected comment block is committed.

### `src/main/java/com/smartContact/model/ErrorMessage.java`
Confidence: 84%
Issues:
  - [info] The migration silently changes the JSON wire format from the Java enum name ("NOT_FOUND") to the numeric code (404). This remains a genuine behavioral difference that could break any consumer depending on the string form; the Expert concedes this cannot be verified from the file scope.

### `src/main/java/com/smartContact/model/User.java`
Confidence: 83%

### `src/main/resources/application.properties`
Confidence: 78%
Issues:
  - [info] The MIGRATION_NOTE references a 'RunMigrations placeholder guidance below' that does not exist in the file; the Target Expert conceded this but the actual code comment has not been changed.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 64%
Issues:
  - [critical] The Target Expert concedes the issue and prescribes the correct fix (rename to end in _test.go), but the response only describes the fix rather than showing it applied. Until the file is actually renamed, `go test` will not discover TestApplicationWiring and the testing import may leak into the production build.

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 85%

### `src/main/java/com/smartContact/repository/UserDao.java`
Confidence: 52%
Issues:
  - [warning] Spring's derived finder throws IncorrectResultSizeDataAccessException when multiple rows match, but the migrated FindByName uses QueryRowContext which silently discards extra rows. The Target Expert concedes this and provides a correct fix, but the fix has not actually been applied to the file — it remains a pending change.
  - [critical] The file contains two `package repository` declarations, which is a hard Go compile error. The Target Expert identified this but the fix is only described, not confirmed applied.

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 79%
Issues:
  - [info] The interface godoc falsely promises a *smarterr.UserNotFoundError on not-found, which the passthrough implementation does not enforce. The Expert conceded this but the fix was only proposed, not confirmed as applied.

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 62%
Issues:
  - [critical] The actual implementations of saveUser, fetchUserList, fetchUserById, deleteUser, updateUser, and getUserNameByName are not present in this file and reside in an unverified userservice.go that was never provided for review.
  - [warning] The line `var _ UserService = (*userService)(nil)` depends on UserService and userService being declared in a file not shown; it cannot compile or be validated in isolation.
  - [warning] The error-on-absence behavior of fetchUserById (throws UserNotFoundException on empty Optional) is not demonstrated as preserved; the Go port risks silently returning nil without an error.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 83%

### `src/main/java/com/smartContact/Controller/UserController.java`
Confidence: 82%
