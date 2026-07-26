// Package model contains the data transfer objects (DTOs) exchanged by the
// SmartContact HTTP API.
//
// MIGRATION_NOTE: The Java source ErrorMessage.java was a Lombok-annotated DTO
// (@Data, @NoArgsConstructor, @AllArgsConstructor) used by the Spring
// @ControllerAdvice layer to serialize error responses (notably the 404 path).
// In Go there is no annotation-based boilerplate generation: the equivalent is
// a plain struct with JSON tags and a constructor. Lombok's generated getters,
// setters, equals, hashCode and toString are unnecessary in idiomatic Go —
// exported fields are accessed directly and value equality is provided by the
// language for comparable structs.
//
// The Java field used Spring's org.springframework.http.HttpStatus enum. Go's
// standard library models HTTP status codes as plain ints (see net/http
// constants such as http.StatusNotFound), so Status is an int here. This keeps
// the struct dependency-free and JSON-serializes cleanly as a numeric status.
package model

// ErrorMessage is the DTO returned in API error responses. It mirrors the
// original Spring ErrorMessage and is the shape written by the writeError
// (404 path) and writeValidationError helpers.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error, e.g.
	// net/http.StatusNotFound. It replaces the Spring HttpStatus enum with a
	// plain int as used throughout the Go standard library.
	Status int `json:"status"`

	// Message is the human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and descriptive message. It is the idiomatic Go replacement for Lombok's
// generated @AllArgsConstructor.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
