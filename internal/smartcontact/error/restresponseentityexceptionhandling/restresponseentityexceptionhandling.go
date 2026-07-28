// Package error provides centralized HTTP error mapping for the smartcontact
// service. It translates domain-level errors (such as a missing user) into
// standardized HTTP responses carrying an ErrorMessage body.
//
// MIGRATION_NOTE: The Java source was a Spring @ControllerAdvice
// (RestResponseEntityExceptionHandling) extending ResponseEntityExceptionHandler.
// Spring's controller-advice mechanism intercepts exceptions thrown anywhere in
// the controller layer via AOP-style proxying. Go has no equivalent runtime
// interception, so this cross-cutting concern is translated into an explicit
// helper (WriteError) plus a chi-compatible middleware (ErrorMapper) that
// callers wire into their handler chain. Handlers funnel their error paths
// through WriteError instead of relying on implicit exception propagation.
package error

import (
	"encoding/json"
	"net/http"

	"github.com/smartContact/internal/smartcontact/model"
)

// WriteError inspects err, maps it to the appropriate HTTP status code, and
// writes a JSON ErrorMessage body to w. It mirrors the Spring exception
// handler: a UserNotFound error becomes HTTP 404, everything else becomes
// HTTP 500. It returns an error if the response body could not be encoded.
//
// MIGRATION_NOTE: The original handler only mapped UserNotFoundException to a
// 404. To keep the error surface complete and production-ready, unrecognized
// errors are mapped to a 500 rather than being silently dropped.
func WriteError(w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	message := err.Error()

	if IsUserNotFound(err) {
		status = http.StatusNotFound
	}

	errorMessage := model.NewErrorMessage(status, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(errorMessage); encErr != nil {
		return encErr
	}
	return nil
}

// ErrorMapper is a chi-compatible middleware placeholder that documents the
// intended wiring point for centralized error handling.
//
// MIGRATION_NOTE: Spring's @ControllerAdvice transparently catches exceptions
// that bubble out of handler methods. Go's net/http has no panic-free way to
// recover a returned error from a standard http.HandlerFunc, because the
// standard handler signature returns nothing. Therefore centralized mapping is
// achieved by handlers calling WriteError directly on their error paths (see
// the handler package). This middleware only provides a panic-recovery safety
// net so an unexpected panic is converted into a 500 ErrorMessage rather than
// crashing the connection. Genuine domain errors should still be reported via
// WriteError from within the handlers.
func ErrorMapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				errorMessage := model.NewErrorMessage(
					http.StatusInternalServerError,
					http.StatusText(http.StatusInternalServerError),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(errorMessage)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
