// Package apperr provides application-level error types and the centralized
// HTTP error-rendering logic for the SmartContact service.
//
// MIGRATION_NOTE: The Java source RestResponseEntityExceptionHandling.java was
// a Spring @ControllerAdvice extending ResponseEntityExceptionHandler. That
// pattern relies on Spring MVC's AOP machinery to intercept exceptions thrown
// anywhere in the controller layer and map them to HTTP responses
// declaratively via @ExceptionHandler.
//
// Go has no equivalent annotation-driven, cross-cutting interceptor. The
// idiomatic translation is an explicit error-rendering function that HTTP
// handlers call in their error branches (typically once, at the top of each
// handler, via a helper). This makes the exception-to-status mapping visible
// and testable rather than magical.
//
// Two behaviours from the original are deliberately preserved:
//
//   - UserNotFoundException (NotFoundError here) maps to HTTP 404 with an
//     ErrorMessage body — the only case the Java @ControllerAdvice explicitly
//     handled.
//   - Everything else falls through to the framework default: a 500. The Java
//     advice did NOT map EmptyResultDataAccessException (raised by JPA's
//     deleteById on a missing id), so that case yielded a 500. WriteError
//     replicates this by treating ErrEmptyResultDelete (and any other
//     unmapped error) as a 500 rather than a 404.
package apperr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smartContact/internal/smartcontact/model"
)

// ErrEmptyResultDelete mirrors Spring Data's EmptyResultDataAccessException,
// which JPA's deleteById raises when the target id does not exist.
//
// MIGRATION_NOTE: The original @ControllerAdvice intentionally did NOT map this
// exception to a 404, so it propagated as a 500. WriteError preserves that
// behaviour by routing this error through the generic (500) branch rather than
// the NotFoundError (404) branch.
var ErrEmptyResultDelete = errors.New("empty result: no rows affected by delete")

// WriteError inspects err and writes the appropriate HTTP status code and a
// JSON ErrorMessage body to w.
//
// It is the Go replacement for the Spring @ControllerAdvice: instead of an
// AOP interceptor, HTTP handlers call WriteError explicitly in their error
// branches.
//
// Mapping rules (preserving the original behaviour exactly):
//
//   - A NotFoundError (or any error wrapping ErrUserNotFound) -> 404.
//   - Every other error, including ErrEmptyResultDelete -> 500.
func WriteError(w http.ResponseWriter, err error) {
	status, msg := classify(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	body := model.NewErrorMessage(status, msg)
	if encErr := json.NewEncoder(w).Encode(body); encErr != nil {
		// The status/headers are already committed; nothing further can be
		// done to inform the client. A production build should log encErr
		// via the injected logger.
		_ = encErr
	}
}

// classify maps an error to its HTTP status code and client-facing message.
// It is separated from WriteError to keep the mapping logic pure and easily
// unit-testable in a table-driven fashion.
func classify(err error) (status int, message string) {
	if err == nil {
		return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
	}

	// UserNotFoundException -> 404. errors.Is walks the wrap chain so both a
	// concrete *NotFoundError and any error wrapping ErrUserNotFound match,
	// mirroring @ExceptionHandler(UserNotFoundException.class).
	var notFound *NotFoundError
	if errors.As(err, &notFound) || errors.Is(err, ErrUserNotFound) {
		return http.StatusNotFound, err.Error()
	}

	// Everything else (including ErrEmptyResultDelete) falls through to the
	// framework-default 500, exactly as the original advice did.
	return http.StatusInternalServerError, err.Error()
}
