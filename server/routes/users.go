package routes

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/users"
)

type UsersRoutes struct {
	Handler users.UsersHandlerInterface
}

/*
Register a new user
API: POST /v1/users/register
*/
func (ur *UsersRoutes) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req users.RegisterUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserName == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	resp, err := ur.Handler.RegisterUser(r.Context(), req)
	if err != nil {
		common.HandleError(err, w)
		return
	}
	slog.Info("User registered: %v", resp.UserId)
	json.NewEncoder(w).Encode(resp)
}

/*
Block a user for the given user ID, op can be block or unblock
API: POST /v1/users/{userId}/{op}
*/
func (ur *UsersRoutes) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	// get the user ID from the URL path
	userId := chi.URLParam(r, "userId")
	if userId == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req users.BlockUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.BlockedUserId == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	op := chi.URLParam(r, "op")
	switch op {
	case "block":
		err = ur.Handler.BlockUser(r.Context(), userId, req)
	case "unblock":
		err = ur.Handler.UnblockUser(r.Context(), userId, req)
	default:
		slog.Error(fmt.Sprintf("Invalid operation: %v", op))
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}

	if err != nil {
		common.HandleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}
