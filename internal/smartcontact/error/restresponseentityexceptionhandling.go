package error

import (
	"context"
	"encoding/json"
	"net/http"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
	"github.com/rs/zerolog/log"
)

// Middleware to handle UserNotFoundException.
func UserNotFoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if err := r.Context().Value("error"); err != nil {
			if repository.IsUserNotFoundError(err) {
				errMessage := model.NewErrorMessage(http.StatusNotFound, "User not found")
				w.WriteHeader(http.StatusNotFound)
				if err := json.NewEncoder(w).Encode(errMessage); err != nil {
					log.Err(err).Msg("failed to write error response")
				}
			}
		}
	})
}