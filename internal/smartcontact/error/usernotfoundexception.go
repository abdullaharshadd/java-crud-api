// Package smarterror provides sentinel errors shared across the SmartContact
// application. It replaces the Java custom-exception hierarchy under
// com.smartContact.error.
//
// MIGRATION_NOTE: The original Java class UserNotFoundException extended the
// checked Exception type and offered the usual set of constructors (empty,
// message, message+cause, cause). Go has no exception hierarchy, so the
// idiomatic replacement is a single sentinel error value that callers compare
// against with errors.Is. The various Java constructors collapse into:
//   - Returning ErrUserNotFound directly for the no-message / message cases.
//   - Wrapping a lower-level cause with fmt.Errorf("...: %w", ErrUserNotFound)
//     (or fmt.Errorf("...: %w", cause) plus errors.Is checks) to preserve the
//     "message + cause" chaining that the Java constructors provided.
//
// The central error-to-HTTP-status handler is expected to map ErrUserNotFound
// to a 404 Not Found response.
//
// NOTE: The package is named smarterror rather than error to avoid shadowing
// the builtin error type/identifier within this package's own scope.
package smarterror

import "errors"

// ErrUserNotFound is the sentinel error signalling that a requested user could
// not be found. It is the Go replacement for the Java
// com.smartContact.error.UserNotFoundException.
//
// Callers should test for it using errors.Is:
//
//	if errors.Is(err, smarterror.ErrUserNotFound) {
//		// respond 404
//	}
//
// To attach context while preserving the sentinel (mirroring the Java
// message+cause constructors) wrap it:
//
//	return fmt.Errorf("lookup user %q: %w", id, smarterror.ErrUserNotFound)
var ErrUserNotFound = errors.New("user not found")
