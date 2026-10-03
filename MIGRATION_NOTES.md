# Migration Notes

**Model self-assessed confidence:** 50% (not a measured result)  
**Build at time of writing:** failed  
**Target unit tests:** not run  
**Behavior compared with the original:** mismatches

The final recommendation is in the pull request description.

---

## What was migrated

- `src/main/java/com/smartContact/model/User.java` → `internal/user/model.go` (78% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/error/UserNotFoundException.java` → `internal/user/errors.go` (89% confidence)
- `src/main/java/com/smartContact/model/ErrorMessage.java` → `internal/httpapi/errors.go` (86% confidence)
- `src/main/java/com/smartContact/repository/UserDao.java` → `internal/user/repository.go` (88% confidence)
- `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java` → `internal/httpapi/errors.go` (74% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/service/UserService.java` → `internal/user/service.go` (89% confidence)
- `src/main/java/com/smartContact/service/UserServiceImp.java` → `internal/user/service.go` (85% confidence) ⚠️ needs review
- `src/main/java/com/smartContact/Controller/UserController.java` → `internal/httpapi/handler.go` (17% confidence) ⚠️ needs review

## Components that could not be automatically migrated

These components require manual implementation. The migrated code contains
`MIGRATION_NOTE` comments at the relevant locations.

### `JpaRepository inherited runtime-generated methods` in `src/main/java/com/smartContact/repository/UserDao.java`
**Reason:** Go has no runtime proxy generation or query derivation from method names. The implementation is synthesized by Spring Data at startup.
**Suggestion:** Manually implement only the methods in use (Save, FindByID, FindByName, Delete, etc.) with sqlx hand-written SQL or gorm. Optionally use sqlc to generate type-safe code from SQL.

### `JPA persistence context semantics (dirty checking, cascades, lazy loading, first-level cache)` in `src/main/java/com/smartContact/repository/UserDao.java`
**Reason:** These behaviors come from Hibernate session management and have no direct equivalent in sqlx. gorm only partially emulates cascades.
**Suggestion:** Rewrite manually: make updates explicit, handle association persistence explicitly within transactions, and replace lazy loading with explicit queries or preloads.

## Observer agent findings

The Observer agent monitored the migration and identified these patterns:

- **After 3 modules:** Confidence is held down mainly by Spring/JPA framework behaviors (id generation, schema init, global exception handling) that have no explicit Go wiring, plus inherent language differences being scored as failures.
- **After 6 modules:** Confidence drops mainly because the target ports each class in isolation and loses Spring/JPA's implicit wiring (schema init, id generation, exception-handler registration). The validator adds noise by listing harmless, unreachable language differences as failed specs.

## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `src/main/java/com/smartContact/model/User.java`
Confidence: 78%
Issues:
  - [warning] EnsureSchema creates the table, but nothing in the target calls it. cmd/server/main.go starts the HTTP server without opening a DB or calling user.EnsureSchema, so the schema is not created at boot right now. Also, user_id has no AUTO_INCREMENT, so inserts only work if the repository takes ids from hibernate_sequence. No repository in the target does this yet.
  - [info] String() includes the password hash. This matches Lombok exactly, but the spec suggests leaving it out to avoid leaking it into logs.

### `src/main/java/com/smartContact/error/RestResponseEntityExceptionHandling.java`
Confidence: 74%
Issues:
  - [info] If UserNotFoundException is created with no message (null), Java returns message null. In Go, NotFoundError.Error() falls back to "User are not available", and a bare ErrNotFound also produces that text. The only place the source throws this exception passes exactly that message, so real usage behaves the same.
  - [info] Spring's handling is now a set of helpers (WriteEmpty, WriteMethodNotAllowed, NotFoundHandler) that each handler and router must call explicitly. Whether they are wired in correctly depends on router.go. There is also no explicit 406 helper.

### `src/main/java/com/smartContact/service/UserServiceImp.java`
Confidence: 85%

### `src/main/java/com/smartContact/Controller/UserController.java`
Confidence: 17%
