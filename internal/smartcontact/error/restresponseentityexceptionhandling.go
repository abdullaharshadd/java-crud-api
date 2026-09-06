package error

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ErrorMessage represents an error response body.
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// UserNotFoundMiddleware is middleware to handle UserNotFoundException.
func UserNotFoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if err := r.Context().Value("error"); err != nil {
			if IsUserNotFoundError(err.(error)) {
				errMessage := &ErrorResponse{Status: http.StatusNotFound, Message: "User not found"}
				w.WriteHeader(http.StatusNotFound)
				if encErr := json.NewEncoder(w).Encode(errMessage); encErr != nil {
					log.Err(encErr).Msg("failed to write error response")
				}
			}
		}
	})
}

// HandleError writes an error response to the http.ResponseWriter.
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	resp := &ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()}
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		log.Err(encErr).Msg("failed to write error response")
	}
}