// Package smartcontacterror defines domain-specific error types for the
// smartContact application.
//
// MIGRATION_NOTE: The Java source was a custom checked exception
// (UserNotFoundException extends Exception) with the standard set of
// Throwable constructors. Go has no exception hierarchy; the idiomatic
// equivalent is a sentinel error plus a typed error that implements the
// error interface and supports errors.Is/errors.As and error wrapping
// via Unwrap(). Callers should return this error (never panic) and
// inspect it with errors.Is(err, ErrUserNotFound) or errors.As.
package smartcontacterror

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error signalling that a requested user
// could not be found. Use errors.Is(err, ErrUserNotFound) to test for it
// regardless of any additional context wrapped around it.
var ErrUserNotFound = errors.New("user not found")

// UserNotFoundError is a typed error indicating that a requested user could
// not be found within the application's domain/business logic.
//
// It optionally carries a human-readable message and a wrapped cause,
// mirroring the message/cause pairs available on the original Java
// exception constructors. It always unwraps to ErrUserNotFound so that
// errors.Is(err, ErrUserNotFound) holds, and to Cause when a cause is set.
type UserNotFoundError struct {
	// Message is an optional human-readable description. When empty, the
	// default "user not found" text is used.
	Message string
	// Cause is the optional underlying error that triggered this one.
	Cause error
}

// NewUserNotFoundError constructs a UserNotFoundError with the given message.
// An empty message falls back to the default ErrUserNotFound text.
//
// MIGRATION_NOTE: This replaces the no-arg and message-only Java
// constructors (UserNotFoundException() and UserNotFoundException(String)).
func NewUserNotFoundError(message string) *UserNotFoundError {
	return &UserNotFoundError{Message: message}
}

// NewUserNotFoundErrorWithCause constructs a UserNotFoundError wrapping the
// given cause together with an optional message.
//
// MIGRATION_NOTE: This replaces the Java constructors that accepted a
// Throwable cause (UserNotFoundException(String, Throwable) and
// UserNotFoundException(Throwable)). The Java protected constructor with
// enableSuppression/writableStackTrace flags has no Go equivalent and is
// intentionally omitted — those flags controlled JVM stack-trace behaviour
// that Go's error model does not expose.
func NewUserNotFoundErrorWithCause(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{Message: message, Cause: cause}
}

// Error implements the error interface.
func (e *UserNotFoundError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = ErrUserNotFound.Error()
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap allows errors.Is/errors.As to traverse into the wrapped cause.
// When no explicit cause is set it exposes ErrUserNotFound so that
// errors.Is(err, ErrUserNotFound) always succeeds for this type.
func (e *UserNotFoundError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return ErrUserNotFound
}

// Is reports whether target is ErrUserNotFound, so that a UserNotFoundError
// with a wrapped cause still matches errors.Is(err, ErrUserNotFound).
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}
