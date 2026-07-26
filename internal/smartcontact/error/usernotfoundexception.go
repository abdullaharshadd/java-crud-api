// Package apperr defines the application's domain error types for the
// SmartContact service.
//
// MIGRATION_NOTE: The Java source UserNotFoundException.java was a custom
// *checked* exception extending java.lang.Exception. Go has no exception
// hierarchy and no checked exceptions; errors are ordinary values that
// implement the built-in `error` interface. The idiomatic translation is a
// concrete error type (NotFoundError) plus a sentinel value (ErrUserNotFound)
// that callers can match with errors.Is / errors.As.
//
// Per the migration rule (Change 13), a missing user MUST always surface as a
// typed NotFoundError produced by NewUserNotFound(id) — it must never be
// wrapped as a bare fmt.Errorf string. This preserves the original 404
// sentinel semantics and lets HTTP handlers map it to a 404 response.
package apperr

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error indicating that a requested user could
// not be located in the system. Use errors.Is(err, ErrUserNotFound) to test
// for this condition regardless of the concrete wrapping error type.
var ErrUserNotFound = errors.New("user not found")

// NotFoundError is the concrete error type returned when a requested user does
// not exist. It carries the identifier that was looked up so callers and
// logging middleware can report useful context, and it wraps ErrUserNotFound
// so that errors.Is(err, ErrUserNotFound) succeeds.
//
// This is the Go equivalent of the original UserNotFoundException. Callers
// should construct it via NewUserNotFound rather than with a struct literal.
type NotFoundError struct {
	// ID is the user identifier that could not be found. It is stored as a
	// string so it can represent numeric IDs, UUIDs, usernames or e-mail
	// addresses without loss.
	ID string

	// Cause is the underlying error that triggered this NotFoundError, if any.
	// It corresponds to the Throwable cause accepted by the original
	// exception's constructors. It may be nil.
	Cause error
}

// NewUserNotFound constructs a NotFoundError for the given user identifier.
// This is the canonical way to signal a missing user throughout the service
// (Change 13 rule) and must be used in place of a bare fmt.Errorf wrap.
func NewUserNotFound(id string) *NotFoundError {
	return &NotFoundError{ID: id}
}

// NewUserNotFoundWithCause constructs a NotFoundError for the given user
// identifier while retaining the underlying cause. This mirrors the Java
// constructor UserNotFoundException(String message, Throwable cause).
func NewUserNotFoundWithCause(id string, cause error) *NotFoundError {
	return &NotFoundError{ID: id, Cause: cause}
}

// Error implements the error interface, returning a human-readable message.
func (e *NotFoundError) Error() string {
	if e.ID == "" {
		if e.Cause != nil {
			return fmt.Sprintf("user not found: %v", e.Cause)
		}
		return "user not found"
	}
	if e.Cause != nil {
		return fmt.Sprintf("user not found: id %q: %v", e.ID, e.Cause)
	}
	return fmt.Sprintf("user not found: id %q", e.ID)
}

// Unwrap allows errors.Is and errors.As to traverse the error chain. It first
// reports the explicit Cause (if present) and otherwise reports the
// ErrUserNotFound sentinel, guaranteeing errors.Is(err, ErrUserNotFound) is
// always true for values of this type.
func (e *NotFoundError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return ErrUserNotFound
}

// Is reports whether the target matches this error. It ensures that a
// NotFoundError always satisfies errors.Is(err, ErrUserNotFound) even when a
// non-nil Cause is set (in which case Unwrap returns the Cause instead of the
// sentinel).
func (e *NotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}
