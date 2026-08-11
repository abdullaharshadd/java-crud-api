package smartcontacterror

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is a sentinel error that can be used with errors.Is to detect
// "user not found" conditions regardless of the specific detail message.
//
// Callers that only care about the category (not the specific user) can write:
//
//	if errors.Is(err, ErrUserNotFound) { ... }
var ErrUserNotFound = errors.New("user not found")

// UserNotFoundError is the typed error signalling that a requested user could
// not be found within the application's domain logic. It is the Go replacement
// for the Java UserNotFoundException.
//
// It carries an optional human-readable message and an optional wrapped cause.
// errors.Is(err, ErrUserNotFound) reports true for any UserNotFoundError, and
// errors.As can be used to recover the concrete value for its Message field.
type UserNotFoundError struct {
	// Message is the detail text describing the failure. It corresponds to the
	// message argument of the Java constructors. It may be empty.
	Message string

	// Cause is the underlying error that triggered this one, if any. It
	// corresponds to the Throwable cause of the Java constructors. It may be
	// nil.
	Cause error
}

// NewUserNotFoundError returns a UserNotFoundError with the given detail
// message. It is the Go equivalent of the UserNotFoundException(String)
// constructor.
func NewUserNotFoundError(message string) *UserNotFoundError {
	return &UserNotFoundError{Message: message}
}

// NewUserNotFoundErrorf returns a UserNotFoundError whose message is built from
// the given printf-style format string and arguments. This is the idiomatic Go
// way to construct a message with interpolated values (e.g. the offending user
// ID).
func NewUserNotFoundErrorf(format string, args ...any) *UserNotFoundError {
	return &UserNotFoundError{Message: fmt.Sprintf(format, args...)}
}

// NewUserNotFoundErrorWithCause returns a UserNotFoundError with the given
// detail message wrapping the supplied cause. It is the Go equivalent of the
// UserNotFoundException(String, Throwable) constructor and enables
// errors.Is/errors.As to traverse the cause chain via Unwrap.
func NewUserNotFoundErrorWithCause(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{Message: message, Cause: cause}
}

// Error implements the error interface. It reports the detail message,
// appending the wrapped cause when one is present.
func (e *UserNotFoundError) Error() string {
	switch {
	case e.Message == "" && e.Cause == nil:
		return "user not found"
	case e.Message == "":
		return fmt.Sprintf("user not found: %v", e.Cause)
	case e.Cause == nil:
		return e.Message
	default:
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
}

// Unwrap returns the wrapped cause, allowing errors.Is and errors.As to walk
// the chain. It corresponds to Java's Throwable.getCause().
func (e *UserNotFoundError) Unwrap() error {
	return e.Cause
}

// Is reports whether target is the ErrUserNotFound sentinel, so that
// errors.Is(err, ErrUserNotFound) matches any UserNotFoundError regardless of
// its message or cause.
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}