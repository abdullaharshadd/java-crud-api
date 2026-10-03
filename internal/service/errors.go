// Package service contains the application's business logic and the domain
// errors it reports to the transport layer.
package service

import "errors"

// UserNotFoundMessage is the detail message the user service attaches when a
// lookup by ID finds nothing. It is surfaced verbatim in the 404 JSON body.
const UserNotFoundMessage = "User are not available"

// ErrUserNotFound signals that a requested user could not be located. Callers
// wrap it with fmt.Errorf("...: %w", ErrUserNotFound) and the HTTP layer maps
// it to 404 via errors.Is. Every *UserNotFoundError also matches it.
var ErrUserNotFound = errors.New(UserNotFoundMessage)

// UserNotFoundError is the Go counterpart of the source's
// UserNotFoundException when a custom detail message and/or underlying cause
// is needed. It matches ErrUserNotFound under errors.Is and exposes its cause
// through Unwrap.
//
// MIGRATION_NOTE: the protected Java constructor controlling suppression and
// stack-trace writability has no Go equivalent (Go errors carry no stack trace
// or suppressed list); it collapses into NewUserNotFoundErrorWithCause.
type UserNotFoundError struct {
	// Message is the detail message; empty means "no message" (Java null).
	Message string
	// Cause is the underlying error, or nil.
	Cause error
}

// NewEmptyUserNotFoundError returns a UserNotFoundError with no detail message
// and no cause.
func NewEmptyUserNotFoundError() *UserNotFoundError {
	return &UserNotFoundError{}
}

// NewUserNotFoundError returns a UserNotFoundError with the given detail
// message and no cause. An empty message means no detail message.
func NewUserNotFoundError(message string) *UserNotFoundError {
	return &UserNotFoundError{Message: message}
}

// NewUserNotFoundErrorWithCause returns a UserNotFoundError with both a detail
// message and an underlying cause.
func NewUserNotFoundErrorWithCause(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{Message: message, Cause: cause}
}

// WrapUserNotFound returns a UserNotFoundError wrapping cause; its message is
// derived from the cause (as Java's Exception(Throwable) does), or empty when
// cause is nil.
func WrapUserNotFound(cause error) *UserNotFoundError {
	e := &UserNotFoundError{Cause: cause}
	if cause != nil {
		e.Message = cause.Error()
	}
	return e
}

// Error implements the error interface and returns the detail message. When no
// message was given, the sentinel's text is returned so the error is never
// rendered as an empty string in logs.
func (e *UserNotFoundError) Error() string {
	if e.Message == "" {
		return ErrUserNotFound.Error()
	}
	return e.Message
}

// DetailMessage returns the raw detail message (empty when none was set),
// mirroring Java's getMessage() for building the 404 response body.
func (e *UserNotFoundError) DetailMessage() string { return e.Message }

// Unwrap returns the underlying cause, if any.
func (e *UserNotFoundError) Unwrap() error { return e.Cause }

// Is reports whether target is ErrUserNotFound, so that every
// UserNotFoundError satisfies errors.Is(err, ErrUserNotFound).
func (e *UserNotFoundError) Is(target error) bool { return target == ErrUserNotFound }

// UserNotFoundDetail extracts the detail message to show in the 404 body for
// any error that matches ErrUserNotFound. It reports false if err does not
// represent a missing user.
func UserNotFoundDetail(err error) (string, bool) {
	if !errors.Is(err, ErrUserNotFound) {
		return "", false
	}
	var unf *UserNotFoundError
	if errors.As(err, &unf) {
		return unf.DetailMessage(), true
	}
	return UserNotFoundMessage, true
}
