// Package model contains data transfer objects (DTOs) used across the
// smartcontact service, including standardized API response payloads.
package model

import "net/http"

// UserNotFoundMessageFormat is the format string used to build the error
// message returned when a user cannot be located by ID.
//
// MIGRATION_NOTE: The exact spacing "User not found with id : %d"
// (space-colon-space) is preserved intentionally, as clients and tests may
// assert against this literal string.
const UserNotFoundMessageFormat = "User not found with id : %d"

// ErrorMessage is a data transfer object representing a standardized error
// response returned to API clients. It carries an HTTP status code and a
// human-readable message.
//
// MIGRATION_NOTE: The original Java field was a Spring HttpStatus enum. In Go
// the idiomatic equivalent is a plain int matching the constants in net/http
// (e.g. http.StatusNotFound). The JSON tags mirror the Lombok-generated
// field names ("status" and "message") so wire compatibility is retained.
type ErrorMessage struct {
	// Status is the HTTP status code associated with this error, using the
	// constants defined in net/http (e.g. http.StatusNotFound).
	Status int `json:"status"`

	// Message is the human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and message. Use net/http status constants for the status argument.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// StatusText returns the textual representation of the ErrorMessage's HTTP
// status code (e.g. "Not Found"). It returns an empty string if the status
// code is unknown to net/http.
func (e ErrorMessage) StatusText() string {
	return http.StatusText(e.Status)
}
