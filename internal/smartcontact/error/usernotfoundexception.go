package apperr

import "fmt"

// UserNotFound is returned when a requested user does not exist.
type UserNotFound struct {
	Message string
}

func (e *UserNotFound) Error() string {
	return fmt.Sprintf("user not found: %s", e.Message)
}

// NewUserNotFound constructs a UserNotFound error with the given message.
func NewUserNotFound(message string) *UserNotFound {
	return &UserNotFound{Message: message}
}