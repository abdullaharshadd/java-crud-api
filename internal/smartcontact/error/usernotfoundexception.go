// Package apperror defines the sentinel errors used across the smartcontact
// service to signal specific domain failure conditions. These errors are
// inspected by the central HTTP error mapper to select the appropriate
// status code (for example, a 404 response for a missing user).
//
// MIGRATION_NOTE: The Java source used the package name "com.smartContact.error"
// and a custom checked exception type (UserNotFoundException extends Exception).
// Go has no checked exceptions and no exception inheritance; the idiomatic
// equivalent is a sentinel error value that callers compare with errors.Is,
// plus a constructor that wraps an underlying cause. The Go package is named
// "apperror" rather than "error" because "error" is a predeclared built-in
// identifier and shadowing it at the package level is poor practice.
package apperror

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error signalling that a requested user could
// not be located within the application's domain logic.
//
// The central error mapper matches this value (via errors.Is) to drive the 404
// response path. It replaces the Java UserNotFoundException's no-arg and
// message-only constructors: use this value directly, or wrap it with
// NewUserNotFoundError to attach context or an underlying cause.
var ErrUserNotFound = errors.New("user not found")

// NewUserNotFoundError returns an error that both carries the supplied
// human-readable message and wraps ErrUserNotFound, so that callers can still
// detect the condition with errors.Is(err, ErrUserNotFound).
//
// This mirrors the Java UserNotFoundException(String message) constructor while
// preserving the sentinel relationship required by the error mapper. If msg is
// empty, ErrUserNotFound is returned unchanged.
//
// MIGRATION_NOTE: The Java constructors that accepted a Throwable cause map to
// NewUserNotFoundErrorf below, which uses %w-style wrapping; the suppression /
// writableStackTrace protected constructor has no Go equivalent and is dropped,
// since Go error values do not model stack-trace suppression semantics.
func NewUserNotFoundError(msg string) error {
	if msg == "" {
		return ErrUserNotFound
	}
	return fmt.Errorf("%s: %w", msg, ErrUserNotFound)
}

// NewUserNotFoundErrorf returns an error wrapping an underlying cause together
// with ErrUserNotFound, allowing errors.Is(err, ErrUserNotFound) to succeed and
// errors.Unwrap to reach the original cause.
//
// This replaces the Java UserNotFoundException(String, Throwable) and
// UserNotFoundException(Throwable) constructors. If cause is nil it behaves
// like NewUserNotFoundError(msg).
func NewUserNotFoundErrorf(msg string, cause error) error {
	if cause == nil {
		return NewUserNotFoundError(msg)
	}
	if msg == "" {
		return fmt.Errorf("%w: %w", ErrUserNotFound, cause)
	}
	return fmt.Errorf("%s: %w: %w", msg, ErrUserNotFound, cause)
}

// IsUserNotFound reports whether err (or any error it wraps) is ErrUserNotFound.
// It is a small convenience wrapper around errors.Is for use by the central
// error mapper and any callers that prefer not to import the sentinel directly.
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
