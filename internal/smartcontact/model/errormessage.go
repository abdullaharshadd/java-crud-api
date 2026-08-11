// Package model contains data transfer objects (DTOs) for the Smart Contact
// service's API layer. It is the Go equivalent of the source project's
// com.smartContact.model package.
//
// MIGRATION_NOTE: The Java source declared ErrorMessage as a Lombok-generated
// DTO (@Data, @NoArgsConstructor, @AllArgsConstructor) carrying a Spring
// HttpStatus and a descriptive message, typically serialized into API error
// responses. In Go there is no Lombok and no Spring HttpStatus enum:
//
//   - Getters/setters are unnecessary; exported struct fields are the idiom.
//   - Spring's HttpStatus maps to net/http status constants (plain int), e.g.
//     http.StatusNotFound. We store the numeric status code so it can be passed
//     directly to http.ResponseWriter.WriteHeader and JSON-encoded cleanly.
//   - JSON field tags replace Jackson's default field-name serialization.
package model

import (
	"net/http"
)

// ErrorMessage is the JSON payload returned to API clients when a request
// fails. It carries the HTTP status code and a human-readable description of
// what went wrong.
//
// It replaces the Java com.smartContact.model.ErrorMessage DTO. The Spring
// HttpStatus field is represented as a plain int status code (use the
// net/http.Status* constants), which serializes cleanly and is directly usable
// with http.ResponseWriter.WriteHeader.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error
	// (e.g. http.StatusNotFound). Use the net/http.Status* constants.
	Status int `json:"status"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage from the given HTTP status code
// and descriptive message. This is the Go equivalent of Lombok's
// @AllArgsConstructor.
//
// Pass a net/http.Status* constant (e.g. http.StatusInternalServerError) as
// status.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// StatusText returns the standard HTTP status text for the ErrorMessage's
// status code (e.g. "Not Found" for 404), or the empty string if the code is
// not a recognized standard status. It is a convenience helper that had no
// direct equivalent in the Java DTO but is useful when logging or rendering
// error responses.
func (e ErrorMessage) StatusText() string {
	return http.StatusText(e.Status)
}
