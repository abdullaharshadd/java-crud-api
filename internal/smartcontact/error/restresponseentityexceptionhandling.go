package error

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"migrated-app/internal/smartcontact/model"
)

// WriteError writes a generic error response with the given HTTP status code
// and message as a JSON-serialised ErrorMessage body. It returns any error
// encountered while marshalling or writing the response so callers can log it.
//
// The Content-Type header and status code are set before the body is written.
func WriteError(w http.ResponseWriter, status int, message string) error {
	errorMessage := model.NewErrorMessage(fmt.Sprintf("%d", status), message)

	body, err := json.Marshal(errorMessage)
	if err != nil {
		// Fall back to a plain 500 if the ErrorMessage itself cannot be
		// marshalled; this should never happen for a well-formed value.
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

// WriteDomainError inspects a domain error and writes the appropriate HTTP
// response, mirroring the exception-to-response mapping performed by the Java
// @ControllerAdvice. It returns true if err matched a known domain error and a
// response was written; false if the error is unrecognised and the caller
// should handle it (e.g. as a generic 500).
//
// Currently the only mapping preserved from the source is:
//   - ErrUserNotFound (UserNotFoundException) -> HTTP 404 Not Found
func WriteDomainError(w http.ResponseWriter, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	if errors.Is(err, ErrUserNotFound) {
		return true, WriteError(w, http.StatusNotFound, err.Error())
	}

	return false, nil
}