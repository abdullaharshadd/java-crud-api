package error

import (
	"errors"
)

type UserNotFoundError struct {
	msg string
}

func (e *UserNotFoundError) Error() string {
	return e.msg
}

func NewUserNotFoundError() error {
	return &UserNotFoundError{msg: errors.New("User not found").Error()}
}