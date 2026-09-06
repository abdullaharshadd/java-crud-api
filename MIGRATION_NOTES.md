# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (90% confidence)
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (20% confidence) ⚠️ needs review
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (76% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling.go` (82% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (79% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/userserviceimptest.go` (79% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (79% confidence) ⚠️ needs review
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `src/main/java/com/smartContact/SmartContactApplication.java`
Confidence: 79%

### `src/main/java/com/smartContact/error/UserNotFoundException.java`
Confidence: 79%

### `src/main/java/com/smartContact/model/User.java`
Confidence: 20%

### `src/main/resources/application.properties`
Confidence: 76%

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 79%

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 82%

### `src/main/java/com/smartContact/repository/UserDao.java`
Confidence: 79%

### `src/main/java/com/smartContact/service/UserService.java`
Confidence: 79%

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 79%

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 79%

### `src/main/java/com/smartContact/Controller/UserController.java`
Confidence: 79%
