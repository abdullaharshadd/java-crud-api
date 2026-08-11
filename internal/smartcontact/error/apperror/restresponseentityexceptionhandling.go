package apperror

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgconn"

	apperr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
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

	var unf *apperr.UserNotFound
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