# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `.mvn/wrapper/maven-wrapper.properties` → `internal/.mvn/wrapper/maven-wrapper.properties.go` (85% confidence)
- `pom.xml` → `internal/pom.xml.go` (37% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/SmartContactApplication.java` → `internal/smartcontact/smartcontactapplication.go` (68% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/smartcontact/error/usernotfoundexception.go` (91% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/smartcontact/model/errormessage.go` (50% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/model/User.java` → `internal/smartcontact/model/user.go` (89% confidence)
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (58% confidence) ⚠️ needs review
- `src/test/java/com/smartContact/SmartContactApplicationTests.java` → `internal/smartcontact/smartcontactapplicationtests.go` (71% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/smartcontact/error/restresponseentityexceptionhandling/restresponseentityexceptionhandling.go` (59% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/smartcontact/repository/userdao.go` (87% confidence)
- `src/main/java/com/smartContact/service/UserService.java` → `internal/smartcontact/service/userservice.go` (85% confidence)
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/smartcontact/service/userserviceimp.go` (88% confidence)
- `src/test/java/com/smartContact/service/UserServiceImpTest.java` → `internal/smartcontact/service/userserviceimptest.go` (2% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/smartcontact/handler/usercontroller.go` (66% confidence) ⚠️ needs review
## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 37%
Issues:
  - [critical] The 'corrected file' is truncated mid-token (ends at 'Mav') and never shows the MySQL driver mapping (github.com/go-sql-driver/mysql) actually applied. The claim of removing the PostgreSQL/$1/$2/RETURNING inventions cannot be verified because the relevant portion of the file is missing.

### `src/main/java/com/smartContact/SmartContactApplication.java`
Confidence: 68%
Issues:
  - [critical] buildRouter() still returns a router wired to a nil *sql.DB, causing fail-late behavior: the process starts and health probe returns 200 while any DB-backed request fails at query time. The Target Expert explicitly concedes the defect and states the fix has not been applied yet ('It should now be applied').

### `src/main/java/com/smartContact/model/ErrorMessage.java`
Confidence: 50%
Issues:
  - [critical] The Target Expert concedes the defect and describes a correct fix (HTTPStatus type with MarshalJSON/UnmarshalJSON producing Spring enum names), but the provided code is truncated mid-map and never shows the actual MarshalJSON/UnmarshalJSON implementations. The fix is not verifiably complete or compilable.

### `src/main/resources/application.properties`
Confidence: 58%
Issues:
  - [critical] The Target Expert correctly concedes the engine swap is an unjustified divergence and describes a precise, correct fix (MySQL defaults, port 3306, user/password root, MySQL DSN format, removal of DBSSLMode, corrected docs). However, this is a described fix — the response does not confirm the code has actually been changed, so the divergence remains present in the current source until the edits are applied.

### `src/test/java/com/smartContact/SmartContactApplicationTests.java`
Confidence: 71%

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 59%
Issues:
  - [warning] The status field's serialized representation (numeric 404 vs Spring enum name 'NOT_FOUND') remains unverified. The expert correctly characterizes the risk and provides a conditional fix but explicitly concedes it cannot be closed from this file; the actual Java wire format and model.ErrorMessage definition were never inspected.

### `src/test/java/com/smartContact/service/UserServiceImpTest.java`
Confidence: 2%
Issues:
  - [critical] The migrated test invents an exported method name (GetUserByName), a *dto.UserResponse return type, and a *string Name field, none of which are derivable from the Java source. The source implies a method returning User with getName() yielding a plain String. As written, got.Name == nil and *got.Name fail to compile if Name is a plain string, and the method call fails if the real signature differs.
  - [critical] Imports use bare internal/smartcontact/... paths lacking the module prefix from go.mod, which will not resolve unless the module is literally named 'internal'. The interior structure was also assumed, and the error package name is unverified.

### `src/main/java/com/smartContact/Controller/UserController.java`
Confidence: 66%
Issues:
  - [warning] validateUser still does not fully replicate the javax.validation constraints declared on model.User. The Target Expert honestly concedes this and proposes a partial structural fix (blank-aware @NotBlank check), but the field-level rules (@Email, @Size, @Pattern, @NotNull on other fields) remain unimplemented pending inspection of model.User.
