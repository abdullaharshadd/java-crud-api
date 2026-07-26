// Package smarterror defines domain-specific error types for the SmartContact
// application. It corresponds to the original Java package
// com.smartContact.error.
//
// MIGRATION_NOTE: Java's checked-exception model (a class extending Exception)
// has no Go equivalent. Idiomatic Go signals failure with error values, not
// exceptions and class hierarchies. The four Java constructors
// (no-arg, message, message+cause, cause-only) collapse into a single Go type
// that implements the error interface and optionally wraps an underlying cause
// (exposed via Unwrap, so errors.Is / errors.As work as expected).
//
// The original class extended Exception (a *checked* exception) specifically so
// that callers were forced to handle the "user not found" case. Go's compiler
// does not enforce this, but returning this error value up the call stack
// preserves the intent, and the HTTP error handler maps it to a 404 response
// (matching Spring's exception-to-HTTP-status mapping).
package smarterror

import "errors"

// ErrUserNotFound is a sentinel error representing the "user could not be found"
// condition. Callers can test for it with errors.Is(err, ErrUserNotFound), and
// the HTTP error handler uses it to return an HTTP 404 (Not Found) response.
//
// This mirrors the role of the original UserNotFoundException in the Spring
// application, where the exception drove a 404 mapping.
var ErrUserNotFound = errors.New("user not found")

// UserNotFoundError is a domain error indicating that a requested user could
// not be found. It carries an optional human-readable message and an optional
// wrapped cause, replacing the four constructors of the original Java
// UserNotFoundException.
//
// It always reports ErrUserNotFound via errors.Is, so handlers can branch on
// the sentinel regardless of how the concrete error was constructed:
//
//	errors.Is(err, ErrUserNotFound) // true
type UserNotFoundError struct {
	// Message is a human-readable description. When empty, the error string
	// falls back to ErrUserNotFound's text.
	Message string
	// Cause is the underlying error that triggered this condition, if any.
	// It may be nil.
	Cause error
}

// NewUserNotFoundError constructs a UserNotFoundError with the given message.
// It corresponds to the Java UserNotFoundException(String message) constructor.
// An empty message is permitted and yields the default "user not found" text.
func NewUserNotFoundError(message string) *UserNotFoundError {
	return &UserNotFoundError{Message: message}
}

// NewUserNotFoundErrorWithCause constructs a UserNotFoundError with a message
// and an underlying cause. It corresponds to the Java
// UserNotFoundException(String message, Throwable cause) constructor. The cause
// is retrievable via errors.Unwrap / errors.As.
func NewUserNotFoundErrorWithCause(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{Message: message, Cause: cause}
}

// Error implements the error interface. It returns the configured message, or
// the default ErrUserNotFound text when no message was supplied.
func (e *UserNotFoundError) Error() string {
	if e.Message == "" {
		return ErrUserNotFound.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped cause, enabling errors.Is and errors.As to inspect
// the chain. It returns nil when no cause was supplied.
func (e *UserNotFoundError) Unwrap() error {
	return e.Cause
}

// Is reports whether target matches this error. It returns true for the
// ErrUserNotFound sentinel so that callers can write
// errors.Is(err, ErrUserNotFound) regardless of how the error was built.
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}
