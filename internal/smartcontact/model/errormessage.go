package model

import "net/http"

// ErrorMessage represents an error response with an HTTP status code and a
// human-readable message, typically returned by exception handlers or REST
// controllers.
//
// MIGRATION_NOTE: The Java source used Spring's HttpStatus enum. In Go we use
// the plain int status codes from net/http (e.g. http.StatusNotFound). The
// unused Java imports (SAXResult, HttpResponse) were dead code and dropped.
type ErrorMessage struct {
	// Status is the HTTP status code (e.g. http.StatusNotFound).
	Status int `json:"status"`
	// Message is the human-readable error description.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and message. This replaces Lombok's @AllArgsConstructor.
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// StatusText returns the standard HTTP status text for the ErrorMessage's
// status code (e.g. "Not Found" for 404). Returns an empty string if the code
// is unknown.
func (e *ErrorMessage) StatusText() string {
	return http.StatusText(e.Status)
}
