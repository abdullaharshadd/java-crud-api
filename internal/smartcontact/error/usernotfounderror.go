// Package error provides domain error types for the Smart Contact service.
package error

import "fmt"

// UserNotFoundError is returned when a requested user does not exist.
// It is the Go equivalent of the Java UserNotFoundException.
type UserNotFoundError struct {
	Message string
}

// Error implements the error interface.
func (e *UserNotFoundError) Error() string {
	return e.Message
}

// NewUserNotFoundErrorf constructs a UserNotFoundError with a formatted message.
func NewUserNotFoundErrorf(format string, args ...interface{}) *UserNotFoundError {
	return &UserNotFoundError{Message: fmt.Sprintf(format, args...)}
}