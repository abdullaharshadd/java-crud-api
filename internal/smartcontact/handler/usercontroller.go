package handler

import (
	"context"
	"net/http"
	"migrated-app/internal/smartcontact/repository"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := 1 // Example ID, replace with actual extraction logic
	repo := repository.GetUserRepository()
	user, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == repository.UserNotFoundError {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	// Proceed with handling the user retrieval
}