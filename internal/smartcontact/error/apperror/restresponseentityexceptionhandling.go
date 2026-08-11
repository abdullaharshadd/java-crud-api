// Package apperror contains the application's error-handling HTTP glue.
//
// MIGRATION_NOTE: The Java source was a Spring @ControllerAdvice
// (RestResponseEntityExceptionHandling) that extended
// ResponseEntityExceptionHandler and declared a single @ExceptionHandler for
// UserNotFoundException, translating it into a 404 ResponseEntity carrying an
// ErrorMessage body. Go has no AOP/annotation-driven interception, so global
// cross-cutting exception handling is realized as an explicit error-writing
// helper that HTTP handlers call (or that a middleware invokes on a returned
// error). This centralizes the exception-to-status mapping the debate notes
// require:
//
//	UserNotFound            -> 404
//	repository.ErrNoRowsDeleted -> 500
//	Postgres duplicate key  -> 500
//	validation errors       -> 400
//	default                 -> 500
//
// The original Spring class only mapped UserNotFoundException->404; the wider
// mapping table comes from the agreed migration plan so a single place owns
// all status decisions.
package apperror

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgconn"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ValidationError marks an error as a client-side validation failure so the
// central handler can map it to HTTP 400. Wrap or return a *ValidationError to
// opt into 400 handling.
//
// MIGRATION_NOTE: Spring resolved validation failures (MethodArgumentNotValid)
// automatically inside ResponseEntityExceptionHandler. Go has no such base
// class, so validation failures must surface as this typed error.
type ValidationError struct {
	// Message is the human-readable validation message returned to the client.
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e == nil {
		return "validation error"
	}
	return e.Message
}

// NewValidationError constructs a *ValidationError with the given message.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// StatusFor inspects err and returns the HTTP status code the application uses
// for it, mirroring the Spring @ExceptionHandler mappings agreed during
// migration.
func StatusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var unf *apperrUserNotFound
	if errors.As(err, &unf) {
		return http.StatusNotFound
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		return http.StatusBadRequest
	}

	if errors.Is(err, repository.ErrUserNotFound) {
		return http.StatusNotFound
	}

	if errors.Is(err, repository.ErrNoRowsDeleted) {
		return http.StatusInternalServerError
	}

	// Postgres duplicate-key violation (SQLSTATE 23505). The Java note referred
	// to MySQL error 1062; the PostgreSQL-equivalent unique_violation is 23505.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

// WriteError serializes err into a model.ErrorMessage JSON body and writes it to
// w with the status code determined by StatusFor. This is the Go replacement
// for the Spring @ControllerAdvice response construction.
func WriteError(w http.ResponseWriter, err error) {
	status := StatusFor(err)

	msg := "internal server error"
	if err != nil {
		msg = err.Error()
	}

	em := model.NewErrorMessage(status, msg)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(em); encErr != nil {
		// The status/header are already committed; fall back to a plain body so
		// the client at least receives something.
		http.Error(w, msg, status)
	}
}

// apperrUserNotFound is a local alias target used only so errors.As can match
// the migrated UserNotFound type from this package without creating an import
// cycle. UserNotFound already lives in this same package
// (usernotfoundexception.go), so we reference it directly.
type apperrUserNotFound = UserNotFound
