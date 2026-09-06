package handler

import (
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/error"
	"net/http"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Query().Get("id")
	user, err := repository.GetUser(ctx, id)
	if err != nil {
		error.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Name))
}