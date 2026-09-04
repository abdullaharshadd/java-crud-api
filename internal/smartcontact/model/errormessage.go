package model

import (
	"net/http"
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

// ToHTTPError converts an ErrorMessage to an HTTP error.
func (e *ErrorMessage) ToHTTPError() error {
	return &httpError{
		errorMessage: e,
	}
}

// httpError wraps an ErrorMessage to implement the error interface.
type httpError struct {
	errorMessage *ErrorMessage
}

func (h *httpError) Error() string {
	return h.errorMessage.Message
}
