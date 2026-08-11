// Package apperr defines application-level error types migrated from the
// Java custom exceptions under com.smartContact.error.
//
// MIGRATION_NOTE: The migration notes specify this type should live in
// internal/apperr as *UserNotFoundError, used by FetchUserById to convert
// an "Optional.empty()" (i.e. a missing row) into a 404 response via
// mapError. The target path given for this file is
// internal/smartcontact/error/usernotfoundexception.go; since "error" is a
// predeclared identifier in Go and an unhelpful package name, this file
// uses the package name "apperr" as directed by the migration notes.
// If the build layout requires the package to match the directory name
// ("error"), rename the package clause accordingly — the type and its
// behavior stay the same.
package apperr

import "fmt"

// UserNotFoundError signals that a requested user could not be found within
// the application's business logic. It is the Go equivalent of the Java
// checked exception com.smartContact.error.UserNotFoundException.
//
// MIGRATION_NOTE: Java modeled this as a checked Exception with four
// constructors (no-arg, message, message+cause, cause). In idiomatic Go
// there is a single value type implementing the error interface, an
// optional wrapped cause exposed via Unwrap (so errors.Is/As work), and
// constructor functions instead of overloaded constructors.
type UserNotFoundError struct {
	// Message is the human-readable description of the error.
	Message string
	// Cause is the underlying error that triggered this one, if any.
	Cause error
}

// DefaultUserNotFoundMessage is the message used when a UserNotFoundError
// is constructed without an explicit one, preserving the parity of the
// Java no-arg constructor which produced an exception with a nil message.
const DefaultUserNotFoundMessage = "user not found"

// Error implements the error interface. It returns the message, appending
// the wrapped cause when present.
func (e *UserNotFoundError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = DefaultUserNotFoundMessage
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap returns the underlying cause so that errors.Is and errors.As can
// traverse the error chain. It mirrors the Java constructor variants that
// accepted a Throwable cause.
func (e *UserNotFoundError) Unwrap() error {
	return e.Cause
}

// NewUserNotFoundError creates a UserNotFoundError with the default
// message. It corresponds to the Java no-arg constructor.
func NewUserNotFoundError() *UserNotFoundError {
	return &UserNotFoundError{Message: DefaultUserNotFoundMessage}
}

// NewUserNotFoundErrorf creates a UserNotFoundError with a formatted
// message. It corresponds to the Java constructor taking a message string.
func NewUserNotFoundErrorf(format string, args ...any) *UserNotFoundError {
	return &UserNotFoundError{Message: fmt.Sprintf(format, args...)}
}

// NewUserNotFoundErrorWithCause creates a UserNotFoundError wrapping an
// underlying cause. It corresponds to the Java constructors that accepted
// a Throwable cause (with or without an accompanying message).
func NewUserNotFoundErrorWithCause(message string, cause error) *UserNotFoundError {
	if message == "" {
		message = DefaultUserNotFoundMessage
	}
	return &UserNotFoundError{Message: message, Cause: cause}
}
