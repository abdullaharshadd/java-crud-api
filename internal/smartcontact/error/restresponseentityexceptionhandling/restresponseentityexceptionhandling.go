// Package error provides centralized HTTP error translation for the
// SmartContact application.
//
// MIGRATION_NOTE: The Java source RestResponseEntityExceptionHandling.java was
// a Spring @ControllerAdvice that extended ResponseEntityExceptionHandler. It
// intercepted UserNotFoundException thrown anywhere in the request pipeline and
// mapped it to a structured 404 response. Go has no AOP / @ControllerAdvice
// mechanism, so the idiomatic replacement is a single WriteError helper that
// each HTTP handler calls explicitly. It inspects the error (using errors.Is
// against migrated sentinels) and writes the appropriate structured JSON
// response.
//
// Divergences from the Java original:
//
//   - There is no base-class inheritance of ResponseEntityExceptionHandler; the
//     framework's default exception handling is replaced by explicit branches
//     plus a generic fallback.
//   - Critically, the original @ControllerAdvice did NOT handle
//     EmptyResultDataAccessException (deleting a missing id bubbles up as a
//     500). We preserve that behaviour: repository.ErrEmptyResultDelete is
//     routed through the generic 500 branch rather than a 404 (Change 12).
//   - Spring's ResponseEntity builder becomes explicit status-code + JSON
//     encoding via encoding/json.
package error

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrUserNotFound is the sentinel error returned by the service layer when a
// requested user does not exist. It is the Go equivalent of the Java
// UserNotFoundException and is matched via errors.Is in WriteError.
var ErrUserNotFound = errors.New("user not found")

// ErrorMessage is the structured error payload returned to API clients. It
// mirrors the Java com.smartContact.model.ErrorMessage record, carrying the
// HTTP status code and a human-readable message.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error.
	Status int `json:"status"`
	// Message is the human-readable error description.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage from an HTTP status code and a
// message. It is the constructor equivalent of the Java ErrorMessage(HttpStatus,
// String) constructor.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{Status: status, Message: message}
}

// WriteError inspects err and writes an appropriate structured JSON error
// response to w. It is the centralized replacement for the Spring
// @ControllerAdvice / @ExceptionHandler mechanism.
//
// Mapping rules:
//
//   - ErrUserNotFound  -> 404 Not Found (mirrors the Java
//     userNotFoundException handler).
//   - anything else     -> 500 Internal Server Error. This deliberately
//     includes repository.ErrEmptyResultDelete, preserving the original
//     @ControllerAdvice behaviour where a missing delete id produced a 500
//     rather than a 404.
//
// WriteError never panics; any JSON encoding failure is silently ignored after
// the status code has been written, matching typical Go HTTP handler behaviour.
func WriteError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, ErrUserNotFound):
		status = http.StatusNotFound
		message = err.Error()
	default:
		// Generic fallback. Note that repository.ErrEmptyResultDelete falls
		// through here on purpose (see MIGRATION_NOTE / Change 12).
		status = http.StatusInternalServerError
		if err != nil {
			message = err.Error()
		} else {
			message = http.StatusText(http.StatusInternalServerError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(NewErrorMessage(status, message))
}
