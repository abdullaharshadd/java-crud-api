package error

import (
	"errors"
)

// UserNotFoundError represents an error indicating that a user was not found.
type UserNotFoundError struct {
	message string
	cause   error
}

// NewUserNotFoundError creates a new UserNotFoundError with the given message and optional cause.
func NewUserNotFoundError(message string, cause error) *UserNotFoundError {
	return &UserNotFoundError{
		message: message,
		cause:   cause,
	}
}

// Error returns the error message for UserNotFoundError.
func (e *UserNotFoundError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

// Unwrap returns the underlying cause of the error.
func (e *UserNotFoundError) Unwrap() error {
	return e.cause
}
