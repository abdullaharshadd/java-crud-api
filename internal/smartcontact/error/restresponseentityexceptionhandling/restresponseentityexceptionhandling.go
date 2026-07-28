// Package error provides HTTP error-mapping utilities for the SmartContact
// service. It replaces the Java Spring @ControllerAdvice global exception
// handler under com.smartContact.error.
//
// MIGRATION_NOTE: The Java source was a @ControllerAdvice class extending
// Spring's ResponseEntityExceptionHandler:
//
//	@ControllerAdvice
//	public class RestResponseEntityExceptionHandling extends ResponseEntityExceptionHandler {
//	    @ExceptionHandler(UserNotFoundException.class)
//	    public ResponseEntity<ErrorMessage> userNotFoundException(...) {
//	        ErrorMessage errorMessage = new ErrorMessage(HttpStatus.NOT_FOUND, exception.getMessage());
//	        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(errorMessage);
//	    }
//	}
//
// Spring registered this handler globally via component scanning and invoked it
// reflectively whenever a controller threw UserNotFoundException. Idiomatic Go
// has no such cross-cutting reflective dispatch. Instead, HTTP handlers return
// (or surface) errors, and a single helper inspects the error taxonomy and
// writes the appropriate JSON response. This file provides that helper
// (WriteError) plus a chi-compatible recovery/error-mapping middleware so the
// mapping remains centralized, matching the intent of the original
// @ControllerAdvice.
package error

import (
	"encoding/json"
	errs "errors"
	"net/http"

	"migrated-app/internal/smartcontact/model"
)

// statusForError maps an application error onto the HTTP status code that best
// represents it. This is the Go equivalent of Spring's @ExceptionHandler
// type-to-status mapping. New error categories should be added here.
func statusForError(err error) int {
	switch {
	case errs.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// WriteError inspects err, selects an appropriate HTTP status code and writes a
// unified ErrorMessage JSON body to w. It is the idiomatic replacement for the
// Spring @ExceptionHandler methods: callers (HTTP handlers) invoke it instead
// of throwing an exception for Spring to intercept.
//
// The original handler mapped UserNotFoundException -> 404 with the exception
// message; that behaviour is preserved via statusForError and the message
// carried by err.
func WriteError(w http.ResponseWriter, err error) {
	status := statusForError(err)

	msg := model.NewErrorMessageFromError(status, err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(msg); encErr != nil {
		// The status line and headers are already committed at this point, so
		// there is nothing meaningful we can send to the client. Fall back to a
		// plain error to make the failure visible in logs/tests.
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

// Recoverer returns a chi-compatible middleware that recovers from panics in
// downstream handlers and converts them into a unified ErrorMessage response.
//
// MIGRATION_NOTE: Spring's ResponseEntityExceptionHandler also provided default
// handling for framework-level exceptions. There is no framework exception
// hierarchy in Go; the closest equivalent is guarding against panics so a
// single unexpected failure does not crash the server or leak a stack trace to
// the client. Business errors should be handled explicitly via WriteError
// rather than by panicking.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch v := rec.(type) {
				case error:
					err = v
				default:
					err = errs.New(http.StatusText(http.StatusInternalServerError))
				}
				WriteError(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
