package error

import (
	"context"
	"net/http"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/handler"
)

// Middleware to handle UserNotFoundException.
func UserNotFoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if err := r.Context().Value("error"); err != nil {
			if _, ok := err.(handler.UserNotFoundError); ok {
				errMessage := model.NewErrorMessage(http.StatusNotFound, err.Error())
				w.WriteHeader(http.StatusNotFound)
				if err := model.ToHTTPError(w, errMessage); err != nil {
					log.Printf("failed to write error response: %v", err)
				}
			}
		}
	})
}