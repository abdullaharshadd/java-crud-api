package error

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smartContact/internal/apperr"
	"github.com/smartContact/internal/model"
)

// RestResponseEntityExceptionHandling is the Go equivalent of the Spring
// @ControllerAdvice class. It maps application errors to HTTP responses,
// mirroring the original ResponseEntityExceptionHandler behavior.
type RestResponseEntityExceptionHandling struct{}

// NewRestResponseEntityExceptionHandling constructs the handler.
func NewRestResponseEntityExceptionHandling() *RestResponseEntityExceptionHandling {
	return &RestResponseEntityExceptionHandling{}
}

// Handle inspects the given error and writes the appropriate HTTP response.
// It is intended to be called from HTTP handlers or recovery middleware.
//
// Equivalent to:
//
//	@ExceptionHandler(UserNotFoundException.class)
//	public ResponseEntity<ErrorMessage> userNotFoundException(...)
func (h *RestResponseEntityExceptionHandling) Handle(w http.ResponseWriter, err error) {
	var unf *apperr.UserNotFoundException
	if errors.As(err, &unf) {
		h.writeError(w, http.StatusNotFound, unf.Error())
		return
	}

	// Fallback for any unmapped error.
	h.writeError(w, http.StatusInternalServerError, err.Error())
}

// Middleware returns an http.Handler wrapper that recovers from panics and
// delegates error rendering to Handle, approximating @ControllerAdvice's
// cross-cutting exception handling.
func (h *RestResponseEntityExceptionHandling) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if err, ok := rec.(error); ok {
					h.Handle(w, err)
					return
				}
				h.writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// writeError builds an ErrorMessage body and writes it with the given status,
// matching ResponseEntity.status(...).body(errorMessage).
func (h *RestResponseEntityExceptionHandling) writeError(w http.ResponseWriter, status int, message string) {
	errorMessage := model.NewErrorMessage(status, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorMessage)
}