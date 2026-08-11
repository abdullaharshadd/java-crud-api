// Package error provides centralized HTTP error mapping for the Smart Contact
// service. It is the Go equivalent of the source project's
// com.smartContact.error package.
//
// MIGRATION_NOTE: The Java source was a Spring @ControllerAdvice class
// (RestResponseEntityExceptionHandling) that extended
// ResponseEntityExceptionHandler and used an @ExceptionHandler method to
// intercept UserNotFoundException across every controller, translating it into
// a 404 ResponseEntity carrying an ErrorMessage body.
//
// Go's net/http has no AOP-style global exception interceptor and no checked
// exceptions. The idiomatic replacement is an explicit mapping function that
// inspects an error value and returns the HTTP status code plus a serializable
// error body. Handlers (or a shared middleware) call MapError once and write
// the result, giving the same "one place decides the response shape" behavior
// without any framework magic.
//
// The Java handler only mapped UserNotFoundException -> 404. Following the
// migration design (CHANGE 18), errors.As(err, &nf) leads the switch, and any
// unrecognized error falls through to a 500 so no handler ever leaks an
// unmapped error to the client.
package error

import (
	errors "errors"
	"net/http"

	"migrated-app/internal/smartcontact/model"
)

// MapError inspects err and returns the HTTP status code and an *ErrorMessage
// describing how it should be reported to an API client.
//
// It is the Go equivalent of the Spring @ControllerAdvice exception handler:
// UserNotFoundError maps to 404 Not Found, and every other (non-nil) error
// maps to 500 Internal Server Error. If err is nil, MapError returns
// (http.StatusOK, nil); callers should treat that as "no error to report".
func MapError(err error) (int, *model.ErrorMessage) {
	if err == nil {
		return http.StatusOK, nil
	}

	var nf *UserNotFoundError
	switch {
	case errors.As(err, &nf):
		return http.StatusNotFound, model.NewErrorMessage(http.StatusNotFound, nf.Error())
	default:
		// MIGRATION_NOTE: The Java class inherited ResponseEntityExceptionHandler,
		// which supplied default handling for a range of Spring MVC exceptions
		// (validation, unreadable body, etc.). There is no Go equivalent to
		// inherit, so all otherwise-unrecognized errors collapse to a generic
		// 500 here. Extend this switch as more domain error types are migrated.
		return http.StatusInternalServerError, model.NewErrorMessage(http.StatusInternalServerError, err.Error())
	}
}

// WriteError maps err to an HTTP status and JSON error body and writes it to w.
// It is a small convenience wrapper around MapError intended for use directly
// from HTTP handlers or a shared error-handling middleware. If err is nil it
// writes nothing and reports false.
func WriteError(w http.ResponseWriter, err error) bool {
	status, msg := MapError(err)
	if msg == nil {
		return false
	}
	model.WriteJSON(w, status, msg)
	return true
}
