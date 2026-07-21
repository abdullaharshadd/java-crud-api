// Package smartcontacterror defines domain-level error values and types
// for the SmartContact service. It is the Go equivalent of the Java
// package com.smartContact.error.
//
// MIGRATION_NOTE: The source file UserNotFoundException.java was a Java
// *checked* exception (it extended Exception, not RuntimeException). In
// Go there is no exception hierarchy and no checked/unchecked
// distinction; errors are ordinary values that implement the built-in
// error interface and are returned explicitly from functions.
//
// The four public Java constructors and the protected 5-argument
// constructor collapse into a small set of idiomatic Go constructs:
//
//   - A sentinel error value (ErrUserNotFound) for the no-argument /
//     simple-message cases, usable with errors.Is.
//   - A UserNotFoundError type carrying an optional message and a
//     wrapped cause, supporting error wrapping via the Unwrap method so
//     that errors.Is / errors.As traverse the chain.
//
// The Java constructor that took enableSuppression / writableStackTrace
// was JVM-specific stack-trace tuning with no Go equivalent and is
// intentionally omitted.
//
// MIGRATION_NOTE: If this error is surfaced over HTTP, a handler or
// middleware should translate it into a 404 Not Found response, mirroring
// what a Spring @ControllerAdvice / @ExceptionHandler would do. That
// translation belongs in the HTTP layer, not here.
package smartcontacterror

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error indicating that a requested user
// could not be found. Callers can test for it with errors.Is, including
// against wrapped UserNotFoundError values.
var ErrUserNotFound = errors.New("user not found")

// UserNotFoundError signals that a requested user could not be located
// within the SmartContact application. It optionally carries a
// descriptive message and an underlying cause, both of which are
// accessible through the standard errors.Is / errors.As mechanisms.
type UserNotFoundError struct {
	// Message is an optional human-readable description. When empty the
	// error text falls back to ErrUserNotFound.
	Message string
	// Cause is the optional underlying error that triggered this one.
	Cause error
}

// NewUserNotFoundError returns a UserNotFoundError with no additional
// message or cause. It is the equivalent of the Java no-argument
// constructor.
func NewUserNotFoundError() *UserNotFoundError {
	return &UserNotFoundError{}
}

// NewUserNotFoundErrorMessage returns a UserNotFoundError carrying the
// given message. It is the equivalent of the Java constructor that took
// a single message string.
func NewUserNotFoundErrorMessage(message string) *UserNotFoundError {
	return &UserNotFoundError{Message: message}
}

// NewUserNotFoundErrorCause returns a UserNotFoundError wrapping the
// given cause. It is the equivalent of the Java constructor that took a
// single Throwable cause.
func NewUserNotFoundErrorCause(cause error) *UserNotFoundError {
	return &UserNotFoundError{Cause: cause}
}

// NewUserNotFoundErrorMessageCause returns a UserNotFoundError carrying
// both a message and a wrapped cause. It is the equivalent of the Java
// constructor that took (message, cause).
func NewUserNotFoundErrorMessageCause(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{Message: message, Cause: cause}
}

// Error implements the error interface, producing a human-readable
// description of the failure.
func (e *UserNotFoundError) Error() string {
	switch {
	case e.Message != "" && e.Cause != nil:
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	case e.Message != "":
		return e.Message
	case e.Cause != nil:
		return fmt.Sprintf("%s: %v", ErrUserNotFound, e.Cause)
	default:
		return ErrUserNotFound.Error()
	}
}

// Unwrap returns the wrapped cause, if any, enabling errors.Is and
// errors.As to traverse the error chain.
func (e *UserNotFoundError) Unwrap() error {
	return e.Cause
}

// Is reports whether the target matches this error. It returns true for
// the ErrUserNotFound sentinel so that callers can write
// errors.Is(err, smartcontacterror.ErrUserNotFound) regardless of
// whether the concrete value is the sentinel or a UserNotFoundError.
func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}
