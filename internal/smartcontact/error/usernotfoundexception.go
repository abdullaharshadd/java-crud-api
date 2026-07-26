// Package apperr defines application-level sentinel errors for the
// SmartContact application. These errors are used by the service layer to
// signal well-known failure conditions (such as a missing user) that callers
// can detect with errors.Is / errors.As and translate into appropriate
// HTTP responses.
//
// MIGRATION_NOTE: The original Java type was
// `com.smartContact.error.UserNotFoundException`, a custom checked exception
// that extended java.lang.Exception with the standard four public
// constructors plus the protected suppression/stack-trace constructor.
//
// Go has no exceptions and no constructor overloading, so the idiomatic
// replacement is:
//   - A sentinel error value (ErrUserNotFound) for identity comparison via
//     errors.Is.
//   - A lightweight error type (UserNotFoundError) that can optionally carry
//     a wrapped cause and a user id, replacing the message/cause constructor
//     overloads. It implements Unwrap so errors.Is/As traverse the chain.
//
// The `enableSuppression` / `writableStackTrace` protected constructor has no
// Go equivalent — Go errors do not carry stack traces by default and there is
// no suppression mechanism — so it is intentionally omitted.
package apperr

import (
	"errors"
	"fmt"

	"github.com/smartContact/internal/smartcontact/model"
)

// ErrUserNotFound is the sentinel error returned by the service layer when a
// requested user cannot be located (for example, when FindByID yields no
// result). Callers should test for it with errors.Is(err, ErrUserNotFound).
var ErrUserNotFound = errors.New("user not found")

// UserNotFoundError is a concrete error type describing a failed user lookup.
// It carries the offending user id and an optional wrapped cause, replacing
// the message/cause constructor overloads of the original Java exception.
//
// It always unwraps to ErrUserNotFound so that
// errors.Is(err, ErrUserNotFound) succeeds regardless of how the error was
// constructed.
type UserNotFoundError struct {
	// UserID is the identifier that could not be resolved. It may be empty
	// when the id is unknown to the caller.
	UserID string
	// Cause is the underlying error that triggered this failure, if any.
	// It is nil when the not-found condition has no lower-level cause.
	Cause error
}

// NewUserNotFoundError builds a UserNotFoundError for the given user id.
// It is the equivalent of the message-only Java constructor and produces an
// error whose message follows model.UserNotFoundMessageFormat.
func NewUserNotFoundError(userID string) *UserNotFoundError {
	return &UserNotFoundError{UserID: userID}
}

// NewUserNotFoundErrorWithCause builds a UserNotFoundError for the given user
// id while preserving the underlying cause. It is the equivalent of the
// (message, cause) Java constructor.
func NewUserNotFoundErrorWithCause(userID string, cause error) *UserNotFoundError {
	return &UserNotFoundError{UserID: userID, Cause: cause}
}

// Error implements the error interface. The message is formatted using
// model.UserNotFoundMessageFormat to keep parity with the original
// application's user-facing wording.
func (e *UserNotFoundError) Error() string {
	msg := fmt.Sprintf(model.UserNotFoundMessageFormat, e.UserID)
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Is reports whether target matches this error. It returns true for the
// ErrUserNotFound sentinel, allowing errors.Is(err, ErrUserNotFound) to
// succeed for any UserNotFoundError value.
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}

// Unwrap exposes the underlying cause so that errors.Is and errors.As can
// traverse the wrapped error chain.
func (e *UserNotFoundError) Unwrap() error {
	return e.Cause
}

// ErrorMessage renders this error as the application's transport-level
// ErrorMessage model, mirroring how a missing-user exception would have been
// surfaced to HTTP clients in the original Spring application.
func (e *UserNotFoundError) ErrorMessage(status int) model.ErrorMessage {
	return model.NewErrorMessage(status, model.StatusText(status), e.Error())
}