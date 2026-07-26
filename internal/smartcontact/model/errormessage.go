// Package model defines data transfer objects (DTOs) for the SmartContact
// application. It corresponds to the original Java package
// com.smartContact.model.
//
// MIGRATION_NOTE: The Java source used Lombok annotations (@Data,
// @NoArgsConstructor, @AllArgsConstructor) to generate boilerplate
// getters/setters/constructors. Go has no equivalent code generation for this;
// instead the struct fields are exported directly and a constructor function
// (NewErrorMessage) is provided. Getters/setters are unnecessary in idiomatic
// Go for a plain DTO.
//
// MIGRATION_NOTE: The Java field `status` was of type
// org.springframework.http.HttpStatus (a Spring enum). There is no direct Go
// equivalent. Since this DTO is serialized in HTTP error responses, the status
// is represented here as an int carrying the numeric HTTP status code (e.g.
// http.StatusNotFound == 404). Callers should populate it using the constants
// in net/http. This is the idiomatic Go representation of an HTTP status.
package model

// ErrorMessage is a data transfer object representing an error response
// payload. It carries the HTTP status code and a human-readable message, and is
// typically serialized to JSON in error responses.
//
// The original Java type used org.springframework.http.HttpStatus for the
// status field; here Status holds the numeric HTTP status code (use the
// constants in net/http, e.g. http.StatusInternalServerError).
type ErrorMessage struct {
	// Status is the numeric HTTP status code associated with the error
	// (for example 404 or 500). Use the constants in net/http to populate it.
	Status int `json:"status"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given HTTP status code
// and message. The status should be a numeric HTTP status code such as those
// defined in net/http (e.g. http.StatusNotFound).
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{
		Status:  status,
		Message: message,
	}
}
