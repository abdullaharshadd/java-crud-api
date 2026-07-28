// Package model contains the data transfer objects (DTOs) exchanged by the
// SmartContact HTTP API. It replaces the Java DTO layer under
// com.smartContact.model.
//
// MIGRATION_NOTE: The Java ErrorMessage class relied on Lombok (@Data,
// @NoArgsConstructor, @AllArgsConstructor) to generate getters/setters,
// constructors, equals/hashCode and toString. In idiomatic Go none of that
// boilerplate is needed: exported struct fields are read/written directly,
// struct values are comparable and printable out of the box, and a
// constructor function replaces the all-args constructor.
//
// The original field type org.springframework.http.HttpStatus is replaced by a
// plain int carrying an HTTP status code (see net/http constants such as
// http.StatusNotFound / http.StatusBadRequest). Per the migration debate this
// single struct is the unified error-response shape used by the error-taxonomy
// middleware for both 400 and 404 responses.
package model

import (
	"errors"
	"net/http"

	smarterror "github.com/smartContact/internal/smartcontact/error"
)

// ErrorMessage is the JSON body returned to clients when a request fails. It
// carries the HTTP status code and a human-readable message describing the
// error. It is the unified error-response shape used by the error-taxonomy
// middleware for both 400 (bad request) and 404 (not found) responses.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error, e.g.
	// http.StatusNotFound or http.StatusBadRequest.
	Status int `json:"status"`

	// Message is the human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage from the given HTTP status code
// and message. It replaces the Lombok @AllArgsConstructor.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// NewErrorMessageFromError maps a Go error to an appropriate ErrorMessage,
// selecting the HTTP status code based on the error taxonomy. Unknown errors
// default to http.StatusBadRequest. This is the intended integration point for
// the error-taxonomy middleware.
func NewErrorMessageFromError(err error) ErrorMessage {
	if err == nil {
		return NewErrorMessage(http.StatusOK, "")
	}

	switch {
	case errors.Is(err, smarterror.ErrUserNotFound):
		return NewErrorMessage(http.StatusNotFound, err.Error())
	default:
		return NewErrorMessage(http.StatusBadRequest, err.Error())
	}
}
