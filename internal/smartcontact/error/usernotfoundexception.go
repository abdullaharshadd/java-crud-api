// Package apperr defines application-specific error types used across the
// smartcontact application. These concrete error types allow HTTP middleware
// to map domain failures to appropriate status codes via errors.As.
//
// MIGRATION_NOTE: The Java source declared UserNotFoundException as a custom
// checked exception (extends Exception). Go has no checked exceptions and no
// exception hierarchy; the idiomatic equivalent is a concrete error type that
// implements the error interface. Java's four constructor overloads collapse
// into a single NewUserNotFound constructor plus an optional wrapped cause,
// exposed through the Unwrap method so errors.Is/errors.As work correctly.
//
// The package path in the debate notes suggested internal/apperr; we honor the
// requested target path (internal/smartcontact/error) but name the package
// "apperr" because "error" is a predeclared identifier in Go and using it as a
// package name is confusing and error-prone.
package apperr

import "fmt"

// UserNotFound signals that a requested user could not be found.
//
// It is used with errors.As in HTTP middleware to produce a 404 response.
type UserNotFound struct {
	// Message is a human-readable description of the failure.
	Message string
	// Cause is the underlying error, if any. It may be nil.
	Cause error
}

// NewUserNotFound returns a *UserNotFound with the given message.
//
// MIGRATION_NOTE: Mirrors the Java UserNotFoundException(String message)
// constructor. Use NewUserNotFoundWithCause to wrap an underlying error.
func NewUserNotFound(message string) *UserNotFound {
	return &UserNotFound{Message: message}
}

// NewUserNotFoundWithCause returns a *UserNotFound wrapping the given cause.
//
// MIGRATION_NOTE: Mirrors the Java UserNotFoundException(String, Throwable)
// constructor. The cause is exposed via Unwrap for errors.Is/errors.As.
func NewUserNotFoundWithCause(message string, cause error) *UserNotFound {
	return &UserNotFound{Message: message, Cause: cause}
}

// Error implements the error interface.
func (e *UserNotFound) Error() string {
	if e.Message == "" {
		if e.Cause != nil {
			return fmt.Sprintf("user not found: %v", e.Cause)
		}
		return "user not found"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause, enabling errors.Is and errors.As to
// traverse the wrapped error chain. It returns nil when there is no cause.
func (e *UserNotFound) Unwrap() error {
	return e.Cause
}
