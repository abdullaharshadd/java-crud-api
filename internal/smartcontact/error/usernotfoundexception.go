// Package apperr defines application-level sentinel errors used across the
// smartcontact service to signal well-known failure conditions such as a
// missing user.
//
// MIGRATION_NOTE: The original Java package was com.smartContact.error and
// contained a checked exception (UserNotFoundException). In idiomatic Go we do
// not model exceptions as types with constructor overloads. Instead we expose a
// sentinel error (ErrNotFound) that callers can match with errors.Is, plus a
// helper constructor that formats a user-specific message and wraps the
// sentinel so both the specific detail and the general category are preserved.
package apperr

import (
	"errors"
	"fmt"

	"github.com/smartContact/internal/smartcontact/model"
)

// ErrNotFound is the sentinel error indicating that a requested user could not
// be located. Service-layer code should return an error that wraps ErrNotFound
// (typically via NewUserNotFound) so callers can detect the condition with
// errors.Is(err, apperr.ErrNotFound).
//
// MIGRATION_NOTE: This replaces the parameterless
// UserNotFoundException() constructor and the checked-exception semantics of
// the original Java class.
var ErrNotFound = errors.New("user not found")

// UserNotFoundError carries the identifier of the user that could not be found
// along with an optional underlying cause. It wraps ErrNotFound so it satisfies
// errors.Is(err, ErrNotFound), and exposes the cause through Unwrap.
//
// MIGRATION_NOTE: The Java class offered four public/protected constructors
// (message, message+cause, cause, and the suppression/stacktrace variant). Go
// has no stack-trace suppression flags, so that constructor is dropped. The
// remaining behaviours are covered by NewUserNotFound and NewUserNotFoundWithCause.
type UserNotFoundError struct {
	// ID is the identifier of the user that could not be located.
	ID int64
	// cause is the optional underlying error, if any.
	cause error
}

// NewUserNotFound returns a UserNotFoundError for the given user ID. The
// resulting error wraps ErrNotFound and formats its message using the shared
// model.UserNotFoundMessageFormat.
func NewUserNotFound(id int64) error {
	return &UserNotFoundError{ID: id}
}

// NewUserNotFoundWithCause returns a UserNotFoundError for the given user ID
// that additionally wraps an underlying cause. The cause is exposed via Unwrap
// so it participates in errors.Is / errors.As chains.
func NewUserNotFoundWithCause(id int64, cause error) error {
	return &UserNotFoundError{ID: id, cause: cause}
}

// Error implements the error interface, producing the standardized
// "user not found" message including the user ID and any underlying cause.
func (e *UserNotFoundError) Error() string {
	msg := fmt.Sprintf(model.UserNotFoundMessageFormat, e.ID)
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.cause)
	}
	return msg
}

// Is reports whether target matches this error. It returns true for
// ErrNotFound so that errors.Is(err, apperr.ErrNotFound) succeeds regardless of
// the concrete ID.
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// Unwrap returns the underlying cause, enabling errors.Is / errors.As to
// traverse the wrapped error chain. If no cause was supplied, it returns
// ErrNotFound so the sentinel remains reachable via Unwrap as well.
func (e *UserNotFoundError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}
	return ErrNotFound
}
