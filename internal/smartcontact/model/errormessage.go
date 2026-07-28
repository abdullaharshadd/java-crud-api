// Package model contains data transfer objects shared across the smartcontact
// service, including the standardized error response shape.
package model

// ErrorMessage is the standardized error response returned by HTTP handlers.
//
// It bundles an HTTP status code together with a human-readable message. In
// the original Java source this used Spring's HttpStatus enum; here we store
// the status as a plain int (matching net/http status constants such as
// http.StatusBadRequest) which is idiomatic for Go HTTP handlers and avoids
// coupling the model to any web framework.
//
// MIGRATION_NOTE (Deviation #2): This single shape unifies the three source
// error formats. The JSON tags below define the wire contract for all error
// responses.
type ErrorMessage struct {
	// Status is the HTTP status code (e.g. http.StatusInternalServerError).
	Status int `json:"status"`
	// Message is the human-readable error description.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and message. It replaces Lombok's @AllArgsConstructor.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
