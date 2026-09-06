package error

import (
	"errors"
)

// UserNotFoundError is returned when a user is not found.
type UserNotFoundError struct {
	msg string
}

func (e *UserNotFoundError) Error() string {
	return e.msg
}

// NewUserNotFoundError creates a new UserNotFoundError.
func NewUserNotFoundError() error {
	return &UserNotFoundError{msg: errors.New("User not found").Error()}
}

// IsUserNotFoundError returns true if the error is a UserNotFoundError.
func IsUserNotFoundError(err error) bool {
	_, ok := err.(*UserNotFoundError)
	return ok
}