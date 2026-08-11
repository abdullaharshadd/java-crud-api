package apperror

import (
	"errors"
	"net/http"

	"migrated-app/internal/smartcontact/model"
)

// validationError is the sentinel/behavioral interface used to classify
// request-validation failures as HTTP 400. Any error whose chain reports
// IsValidationError() == true (or that satisfies this interface directly) is
// treated as a client-side validation problem.
//
// MIGRATION_NOTE: the source only handled UserNotFoundException; the 400
// branch is required by the migration notes (CHANGE 15). We detect it via a
// small behavioral interface so any package can opt in without importing a
// concrete type.
type validationError interface {
	error
	IsValidationError() bool
}

// MapError inspects err and returns the HTTP status code and a populated
// model.ErrorMessage describing the failure. It is the Go replacement for the
// Spring @ControllerAdvice exception handler: callers at the HTTP boundary
// invoke MapError and write the returned status and body.
//
//   - *UserNotFoundError            -> 404 Not Found
//   - errors satisfying validationError -> 400 Bad Request
//   - any other non-nil error       -> 500 Internal Server Error
//
// If err is nil, MapError returns http.StatusInternalServerError with a
// generic message, since callers should not invoke it without an error.
func MapError(err error) (int, model.ErrorMessage) {
	if err == nil {
		return http.StatusInternalServerError, model.NewErrorMessage(
			model.HTTPStatus(http.StatusInternalServerError),
			http.StatusText(http.StatusInternalServerError),
		)
	}

	var notFound *UserNotFoundError
	if errors.As(err, &notFound) {
		return http.StatusNotFound, model.NewErrorMessage(
			model.HTTPStatus(http.StatusNotFound),
			notFound.Error(),
		)
	}

	var valErr validationError
	if errors.As(err, &valErr) {
		return http.StatusBadRequest, model.NewErrorMessage(
			model.HTTPStatus(http.StatusBadRequest),
			valErr.Error(),
		)
	}

	return http.StatusInternalServerError, model.NewErrorMessage(
		model.HTTPStatus(http.StatusInternalServerError),
		err.Error(),
	)
}