package error

import (
	"encoding/json"
	"errors"
	"net/http"

	"migrated-app/internal/smartcontact/model"
)

// ErrUserNotFound is the sentinel error for a user-not-found condition.
// It mirrors the Java UserNotFoundException and is used by errors.Is checks.
var ErrUserNotFound = errors.New("user not found")

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