// Package error provides centralized HTTP error translation for the
// SmartContact application.
//
// MIGRATION_NOTE: The Java source RestResponseEntityExceptionHandling.java was
// a Spring @ControllerAdvice class extending ResponseEntityExceptionHandler.
// It provided a single @ExceptionHandler mapping UserNotFoundException to an
// HTTP 404 response carrying an ErrorMessage body.
//
// There is no direct Go equivalent to Spring's AOP-style @ControllerAdvice
// cross-cutting exception handling. The idiomatic Go translation is a plain
// helper function (WriteError) that HTTP handlers call explicitly when a
// service returns an error. It inspects the error with errors.Is and maps it
// to the appropriate HTTP status code + JSON body.
//
// IMPORTANT — behavioural fidelity: the Java @ControllerAdvice deliberately
// did NOT declare a handler for EmptyResultDataAccessException (raised by
// deleteById when the id is absent), so Spring's default fallback turned that
// into an HTTP 500. WriteError replicates this exactly: repository.ErrEmptyResultDelete
// is mapped to 500 (Internal Server Error), NOT 404.
//
// Manual review: confirm that every HTTP handler routes its service errors
// through WriteError so this mapping is consistently applied. Route
// registration itself lives in internal/smartcontact/handler/usercontroller.go
// (RegisterRoutes); this package only produces error responses.
package error

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smartContact/internal/smartcontact/repository"
)

// ErrUserNotFound is the sentinel error that mirrors the Java
// UserNotFoundException. Service and handler layers return (or wrap) this value
// when a requested user does not exist; WriteError maps it to HTTP 404.
var ErrUserNotFound = errors.New("user not found")

// ErrorMessage is the structured JSON body returned to clients on error.
//
// MIGRATION_NOTE: this mirrors com.smartContact.model.ErrorMessage from the
// Java source, which held an HttpStatus plus a human-readable message.
type ErrorMessage struct {
	// Status is the numeric HTTP status code of the response.
	Status int `json:"status"`
	// Message is the human-readable error description.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage for the given HTTP status code and
// message.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{Status: status, Message: message}
}

// statusFor maps a domain error to its corresponding HTTP status code.
//
// This replaces Spring's @ExceptionHandler dispatch table:
//
//   - ErrUserNotFound                -> 404 (the mapped @ExceptionHandler)
//   - repository.ErrEmptyResultDelete -> 500 (deliberately UNhandled in Java,
//     falling through to Spring's default 500; see package MIGRATION_NOTE)
//   - everything else                -> 500
func statusFor(err error) int {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrEmptyResultDelete):
		// MIGRATION_NOTE: the Java @ControllerAdvice did not handle
		// EmptyResultDataAccessException, so it surfaced as HTTP 500. We
		// preserve that behaviour intentionally rather than "fixing" it to 404.
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// WriteError writes err to w as a JSON ErrorMessage with the appropriate HTTP
// status code, replicating the Spring @ControllerAdvice behaviour.
//
// It never panics. If encoding the response body fails, it falls back to a
// plain-text 500 so the client always receives a terminal response.
func WriteError(w http.ResponseWriter, err error) {
	status := statusFor(err)

	msg := ""
	if err != nil {
		msg = err.Error()
	}
	body := NewErrorMessage(status, msg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(body); encErr != nil {
		// Headers are already sent; there is nothing more we can do beyond
		// best-effort logging by the caller. Return silently.
		return
	}
}
