package model

import (
	"net/http"
)

// HTTPStatus represents an HTTP status code that serializes to its Spring
// HttpStatus enum name (e.g. "NOT_FOUND") for JSON parity with the source
// Java service, which used Jackson to serialize org.springframework.http.HttpStatus.
type HTTPStatus int

// httpStatusNames maps HTTP status codes to their Spring HttpStatus enum
// names. Only the subset actually produced by the service's mapError logic
// needs to be present, but the common codes are included for safety.
//
// MIGRATION_NOTE: Spring's HttpStatus enum name is NOT identical to Go's
// http.StatusText() (e.g. Go returns "Not Found", Spring's enum name is
// "NOT_FOUND"). We therefore maintain an explicit table to preserve the
// exact wire format the original API clients expect.
var httpStatusNames = map[HTTPStatus]string{
	http.StatusOK:                  "OK",
	http.StatusCreated:             "CREATED",
	http.StatusAccepted:            "ACCEPTED",
	http.StatusNoContent:           "NO_CONTENT",
	http.StatusBadRequest:          "BAD_REQUEST",
	http.StatusUnauthorized:        "UNAUTHORIZED",
	http.StatusForbidden:           "FORBIDDEN",
	http.StatusNotFound:            "NOT_FOUND",
	http.StatusMethodNotAllowed:    "METHOD_NOT_ALLOWED",
	http.StatusConflict:            "CONFLICT",
	http.StatusUnprocessableEntity: "UNPROCESSABLE_ENTITY",
	http.StatusInternalServerError: "INTERNAL_SERVER_ERROR",
	http.StatusNotImplemented:      "NOT_IMPLEMENTED",
	http.StatusBadGateway:          "BAD_GATEWAY",
	http.StatusServiceUnavailable:  "SERVICE_UNAVAILABLE",
}

// Name returns the Spring-style enum name for the status (e.g. "NOT_FOUND").
// If the status is not known, the numeric code rendered as a string is
// returned via strconv-free fallback in MarshalJSON.
func (s HTTPStatus) Name() (string, bool) {
	name, ok := httpStatusNames[s]
	return name, ok
}

// Code returns the underlying integer HTTP status code.
func (s HTTPStatus) Code() int {
	return int(s)
}

// MarshalJSON serializes the status as its Spring HttpStatus enum name to
// preserve JSON parity with the original Java service. Unknown codes fall
// back to their numeric string form.
func (s HTTPStatus) MarshalJSON() ([]byte, error) {
	if name, ok := httpStatusNames[s]; ok {
		return []byte(`"` + name + `"`), nil
	}
	// Fallback: render the numeric code as a JSON string so we never
	// silently emit an invalid/empty value.
	return []byte(`"` + http.StatusText(int(s)) + `"`), nil
}

// ErrorMessage is the standardized error payload returned to API clients.
// It mirrors the source com.smartContact.model.ErrorMessage DTO.
type ErrorMessage struct {
	Status  HTTPStatus `json:"status"`
	Message string     `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given status and
// message. This replaces Lombok's @AllArgsConstructor.
func NewErrorMessage(status HTTPStatus, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
