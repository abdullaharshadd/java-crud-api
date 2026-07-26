// Package error provides centralized HTTP error handling for the SmartContact
// application.
//
// MIGRATION_NOTE: The original Java type was a Spring
// `@ControllerAdvice`-annotated class extending
// `ResponseEntityExceptionHandler`. Spring used AOP-style interception to
// globally catch exceptions thrown by any `@Controller`/`@RestController` and
// translate them into structured `ResponseEntity` responses. The
// `@ExceptionHandler(UserNotFoundException.class)` method mapped a
// UserNotFoundException to a 404 response carrying an ErrorMessage body.
//
// Go has no equivalent annotation-driven, AOP-based exception interception. The
// idiomatic replacement is HTTP middleware plus an explicit error-to-response
// mapping helper. Handlers return errors up the stack (or record them on the
// request context/response), and this middleware/helper inspects the error and
// writes the appropriate status code and JSON body.
//
// Mapping rules (preserving the original behavior):
//   - A UserNotFoundError (errors.Is ErrUserNotFound) -> 404 Not Found.
//   - Any other (unrecognized) error -> 500 Internal Server Error. This mirrors
//     Spring's default handling of unhandled exceptions such as the
//     EmptyResultDataAccessException raised by deleting a missing id.
package error

import (
	"encoding/json"
	"errors"
	"net/http"

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// ErrorResponder maps a handler error to an HTTP status code and a structured
// ErrorMessage body, then writes it to the ResponseWriter. It is the Go
// equivalent of the Spring @ControllerAdvice exception handler.
type ErrorResponder struct{}

// NewErrorResponder constructs an ErrorResponder.
func NewErrorResponder() *ErrorResponder {
	return &ErrorResponder{}
}

// StatusAndMessage inspects err and returns the HTTP status code and the
// ErrorMessage that should be sent to the client.
//
// A UserNotFoundError yields 404 Not Found (replicating the original
// @ExceptionHandler(UserNotFoundException.class) branch). Any other error
// yields 500 Internal Server Error, replicating Spring's default handling of
// otherwise-unhandled exceptions.
func (r *ErrorResponder) StatusAndMessage(err error) (int, model.ErrorMessage) {
	if err == nil {
		return http.StatusOK, model.NewErrorMessage(http.StatusOK, "")
	}

	var notFound *smartcontacterror.UserNotFoundError
	if errors.As(err, &notFound) || errors.Is(err, smartcontacterror.ErrUserNotFound) {
		return http.StatusNotFound, model.NewErrorMessage(http.StatusNotFound, err.Error())
	}

	return http.StatusInternalServerError, model.NewErrorMessage(http.StatusInternalServerError, err.Error())
}

// Write serializes the mapped ErrorMessage for err to w as JSON with the
// appropriate HTTP status code. It is safe to call with a nil error, in which
// case nothing is written and false is returned.
func (r *ErrorResponder) Write(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	status, msg := r.StatusAndMessage(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// If encoding fails there is nothing meaningful left to send since the
	// status line has already been written; the error is intentionally ignored
	// after a best-effort attempt.
	_ = json.NewEncoder(w).Encode(msg)
	return true
}

// HandlerFunc mirrors an HTTP handler that may return an error, allowing errors
// to propagate up to centralized handling rather than each handler writing its
// own error response.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Middleware adapts an error-returning HandlerFunc into a standard
// http.Handler. Any error returned by next is translated into a structured
// response via the ErrorResponder.
//
// MIGRATION_NOTE: This is the Go replacement for Spring's global
// @ControllerAdvice interception. Wrap route handlers with this middleware at
// composition time to obtain the same cross-cutting error-mapping behavior.
func (r *ErrorResponder) Middleware(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := next(w, req); err != nil {
			r.Write(w, err)
		}
	})
}
