// Package model contains data transfer objects (DTOs) for the
// SmartContact service. It is the Go equivalent of the Java package
// com.smartContact.model.
//
// MIGRATION_NOTE: The source ErrorMessage.java was a Lombok-generated
// DTO (@Data, @NoArgsConstructor, @AllArgsConstructor) carrying an HTTP
// status and a descriptive message, typically serialized into API error
// responses.
//
// The Lombok-generated boilerplate (getters, setters, equals, hashCode,
// toString, and the two constructors) collapses into idiomatic Go:
//
//   - A plain struct with exported fields and JSON tags replaces the
//     getters/setters and enables direct encoding/json marshalling.
//   - A NewErrorMessage constructor replaces the @AllArgsConstructor.
//   - The zero value of the struct replaces the @NoArgsConstructor.
//
// The Spring HttpStatus enum field is replaced by a plain int holding
// the HTTP status code (matching the values in net/http, e.g.
// http.StatusNotFound). This keeps the type serialization-friendly and
// avoids pulling in a Spring-style enum abstraction.
//
// The unused Java imports (javax.xml.transform.sax.SAXResult and
// java.net.http.HttpResponse) carried no behaviour and are dropped.
package model

// ErrorMessage is a data transfer object representing an API error
// response payload. It combines an HTTP status code with a descriptive
// message string.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error. Use the
	// constants from net/http (e.g. http.StatusNotFound) for the value.
	Status int `json:"status"`

	// Message is the human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage from the given HTTP status
// code and descriptive message. It is the Go equivalent of the Lombok
// @AllArgsConstructor.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
