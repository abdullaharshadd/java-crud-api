// Package model contains data transfer objects used across the smartcontact
// application, including the error response envelope returned by middleware.
package model

// ErrorMessage represents an error response payload returned to HTTP clients.
// It holds an HTTP status code and a human-readable descriptive message.
//
// MIGRATION_NOTE: The Java source used Spring's HttpStatus enum for the status
// field. In idiomatic Go we store the status as a plain int (matching the
// values in net/http, e.g. http.StatusBadRequest). This keeps the model free
// of framework coupling and lets the JSON payload carry the numeric status
// directly. Lombok's @Data / @NoArgsConstructor / @AllArgsConstructor
// boilerplate is replaced by the exported struct fields plus the NewErrorMessage
// constructor below.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error
	// (e.g. http.StatusBadRequest, http.StatusInternalServerError).
	Status int `json:"status"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and descriptive message.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
