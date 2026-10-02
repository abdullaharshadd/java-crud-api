// Package apperr defines the application's sentinel errors used to signal
// well-known business-logic conditions across the service and HTTP layers.
//
// MIGRATION_NOTE: The source defined UserNotFoundException as a custom checked
// exception (extending java.lang.Exception) with the full five-constructor
// boilerplate (no-arg, message, message+cause, cause, and the protected
// suppression/stack-trace constructor). Go has no checked-exception hierarchy
// and no stack-trace/suppression knobs on errors, so that machinery does not
// translate 1:1. In idiomatic Go the single thing that code actually relied on
// — "the requested user was not found" — is a sentinel error value compared
// with errors.Is. UserServiceImp only ever threw it with the message
// "User are not available", and the exception handler turned it into a 404
// with that message. Both of those behaviors are preserved by ErrUserNotFound.
//
// MIGRATION_NOTE: The message+cause / cause-only constructors have no distinct
// consumer in the source (the service always used the message-only form). If a
// future caller needs to attach a cause, wrap with fmt.Errorf("...: %w", err)
// and still match via errors.Is(err, ErrUserNotFound).
package apperr

import "errors"

// ErrUserNotFound signals that a requested user could not be located.
//
// Its message text ("User are not available") is intentionally identical to
// the string the source service passed to new UserNotFoundException(...), so
// the 404 response body is byte-for-byte preserved. Match it with
// errors.Is(err, ErrUserNotFound).
var ErrUserNotFound = errors.New("User are not available")
