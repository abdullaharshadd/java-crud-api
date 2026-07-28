// Package error provides application-specific error types and the HTTP
// error-handling middleware for the smartContact application.
//
// MIGRATION_NOTE: The Java source (RestResponseEntityExceptionHandling) was a
// Spring @ControllerAdvice class extending ResponseEntityExceptionHandler. It
// centralized exception-to-HTTP-response mapping across all controllers via
// AOP-style interception. Go has no equivalent runtime interception or class
// inheritance, so the idiomatic replacement is an HTTP middleware plus an
// explicit error-writing helper. Handlers signal a UserNotFoundError up the
// call stack (via panic-recovery here, or by returning it and calling
// WriteError directly) and this code translates it into a structured 404
// response, mirroring the original @ExceptionHandler(UserNotFoundException).
//
// MIGRATION_NOTE: The inherited ResponseEntityExceptionHandler behaviour
// (default handling of standard Spring MVC exceptions such as validation and
// message-not-readable errors) has no direct Go analogue. Only the explicit
// UserNotFoundException mapping is preserved; other errors fall through to a
// generic 500 response. Extend WriteError as additional error types are
// migrated.
package error

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smartContact/internal/smartcontact/model"
)

// WriteError inspects err, selects the appropriate HTTP status code, and writes
// a structured model.ErrorMessage JSON body to w. It returns the status code
// that was written so callers may log it.
//
// It replaces the Spring @ExceptionHandler dispatch: a *UserNotFoundError (or
// any error wrapping ErrUserNotFound) produces a 404 response carrying the
// error message, exactly as the Java userNotFoundException handler did. Any
// other error produces a generic 500 response.
func WriteError(w http.ResponseWriter, err error) int {
	if err == nil {
		return http.StatusOK
	}

	status := http.StatusInternalServerError
	if errors.Is(err, ErrUserNotFound) {
		status = http.StatusNotFound
	}

	msg := model.NewErrorMessage(status, err.Error())
	writeJSON(w, status, msg)
	return status
}

// writeJSON serializes body as JSON and writes it to w with the given status
// code. Any encoding failure is written as a plain 500 so the client always
// receives a response.
func writeJSON(w http.ResponseWriter, status int, body *model.ErrorMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RecoverMiddleware returns HTTP middleware that recovers from panics carrying
// an error value and translates them into structured error responses via
// WriteError. This provides the cross-cutting, controller-wide error handling
// that the Java @ControllerAdvice offered without requiring every handler to
// translate errors itself.
//
// MIGRATION_NOTE: Panicking to signal errors is not idiomatic Go for ordinary
// control flow; prefer returning errors and calling WriteError explicitly.
// This middleware exists only as a safety net so an escaped UserNotFoundError
// (or any error) still yields the correct structured response rather than a
// bare 500 with no body.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if err, ok := rec.(error); ok {
					WriteError(w, err)
					return
				}
				// Re-panic non-error values so the process fails loudly
				// rather than silently swallowing a programming bug.
				panic(rec)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
