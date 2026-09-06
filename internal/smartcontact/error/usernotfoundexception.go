package error

import (
	"errors"
)

type UserNotFoundError struct {
	err error
}

func (e *UserNotFoundError) Error() string {
	return e.err.Error()
}

func NewUserNotFoundError() error {
	return &UserNotFoundError{err: errors.New("User not found")}
}