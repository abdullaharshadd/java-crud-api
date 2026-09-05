package handler

import (
	"context"
	"net/http"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/model"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := 1 // Example ID
	userRepo := repository.NewUserRepository()

	user, err := userRepo.GetUser(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}