package model

import (
	"errors"
)

// ErrorMessage represents an error message with an HTTP status and a message.
type ErrorMessage struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewErrorMessage creates a new ErrorMessage with the given status and message.
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// FromHTTPError creates an ErrorMessage from an HTTP status and an error.
func FromHTTPError(status int, err error) *ErrorMessage {
	if err == nil {
		return nil
	}
	return NewErrorMessage(status, err.Error())
}

// ToHTTPError converts an ErrorMessage to an error.
func (e *ErrorMessage) ToHTTPError() error {
	return errors.New(e.Message)
}