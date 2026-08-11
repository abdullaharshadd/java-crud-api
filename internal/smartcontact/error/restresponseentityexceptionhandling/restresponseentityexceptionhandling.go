package error

import (
	"encoding/json"
	"errors"
	"net/http"

	smartcontacterror "migrated-app/internal/smartcontact/error"
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

	var nf *smartcontacterror.UserNotFoundError
	switch {
	case errors.As(err, &nf):
		msg := model.NewErrorMessage(http.StatusNotFound, nf.Error())
		return http.StatusNotFound, &msg
	default:
		// MIGRATION_NOTE: The Java class inherited ResponseEntityExceptionHandler,
		// which supplied default handling for a range of Spring MVC exceptions
		// (validation, unreadable body, etc.). There is no Go equivalent to
		// inherit, so all otherwise-unrecognized errors collapse to a generic
		// 500 here. Extend this switch as more domain error types are migrated.
		msg := model.NewErrorMessage(http.StatusInternalServerError, err.Error())
		return http.StatusInternalServerError, &msg
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(msg)
	return true
}