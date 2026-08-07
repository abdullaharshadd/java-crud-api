// Package error defines domain-level sentinel errors for the smartcontact
// application.
//
// MIGRATION_NOTE: The Java source defined a custom checked exception,
// UserNotFoundException, extending Exception. Idiomatic Go does not use
// exceptions; instead this concern is modeled as a sentinel error value
// (ErrUserNotFound) that callers can compare against with errors.Is. This
// mirrors the standard library's sql.ErrNoRows pattern.
//
// Per the migration debate notes, this maps to a "not found" error used
// exclusively by the by-id lookup path where 404 behavior is confirmed.
//
// MIGRATION_NOTE: The Java package was literally named "error". Go allows a
// package named "error" (it is not a reserved keyword; only the builtin
// identifier `error` is predeclared), but importing it may shadow the builtin
// `error` type at call sites. Consumers should import this package with an
// alias, e.g. `smerr "github.com/example/.../internal/smartcontact/error"`,
// to avoid confusion. Renaming the package (e.g. to `domainerr`) is
// recommended and worth manual review.
package error

import (
	stderrors "errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error returned when a requested user cannot
// be located. Callers should test for it with errors.Is(err, ErrUserNotFound)
// rather than comparing error strings.
//
// It replaces the Java UserNotFoundException. Handlers that receive this error
// (or any error wrapping it) should respond with HTTP 404 Not Found.
var ErrUserNotFound = stderrors.New("user not found")

// NewUserNotFound returns an error that wraps ErrUserNotFound with additional
// context, typically identifying the user that could not be found. The
// returned error unwraps to ErrUserNotFound, so errors.Is(err, ErrUserNotFound)
// reports true.
//
// It corresponds to the Java UserNotFoundException(String message)
// constructor.
func NewUserNotFound(message string) error {
	return fmt.Errorf("%s: %w", message, ErrUserNotFound)
}

// WrapUserNotFound returns an error that wraps both a descriptive message and
// an underlying cause, while still unwrapping to ErrUserNotFound. Use this when
// a lower-level error (the cause) explains why the user could not be located.
//
// It corresponds to the Java UserNotFoundException(String message, Throwable
// cause) constructor. Note that Go's error chain permits only a single wrapped
// error per %w verb, so the sentinel is the primary wrapped value and the
// cause is included as context; if callers need errors.Is to match the cause
// as well, use errors.Join at the call site.
func WrapUserNotFound(message string, cause error) error {
	if cause == nil {
		return NewUserNotFound(message)
	}
	return fmt.Errorf("%s: %v: %w", message, cause, ErrUserNotFound)
}
