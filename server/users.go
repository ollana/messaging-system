package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/db"
)

type RegisterUserRequest struct {
	UserName string `json:"UserName"`
}
type RegisterUserResponse struct {
	UserId   string `json:"UserId"`
	UserName string `json:"UserName"`
}

/*
Register a new user
API: POST /v1/users/register
*/

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req RegisterUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserName == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	resp, err := registerUser(r.Context(), req)
	if err != nil {
		common.HandleError(err, w)
		return
	}
	slog.Info("User registered: %v", resp.UserId)
	json.NewEncoder(w).Encode(resp)
}
func registerUser(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error) {
	// generate a new user ID with UUID
	userId := fmt.Sprintf("user-%s", uuid.New().String())

	// store the user in the database
	user := db.User{
		UserId:       userId,
		UserName:     req.UserName,
		BlockedUsers: make(map[string]bool),
	}
	err := dbClient.StoreUser(ctx, user)
	if err != nil {
		slog.Error(fmt.Sprintf("Error storing user: %v", err))
		return nil, &common.InternalServerError{Message: "Error storing user"}
	}
	// return the user ID and name in the response
	resp := RegisterUserResponse{
		UserId:   userId,
		UserName: req.UserName,
	}
	return &resp, nil
}

type BlockUserRequest struct {
	BlockedUserId string `json:"BlockedUserId"`
}

/*
Block a user for the given user ID, op can be block or unblock
API: POST /v1/users/{userId}/{op}
*/
func blockUserHandler(w http.ResponseWriter, r *http.Request) {
	// get the user ID from the URL path
	userId := chi.URLParam(r, "userId")
	if userId == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req BlockUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.BlockedUserId == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	op := chi.URLParam(r, "op")
	switch op {
	case "block":
		err = blockUser(r.Context(), userId, req)
	case "unblock":
		err = unblockUser(r.Context(), userId, req)
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

func unblockUser(ctx context.Context, userId string, req BlockUserRequest) error {

	user, err := dbClient.GetUser(ctx, userId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", userId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// get blocked user
	blockedUser, err := dbClient.GetUser(ctx, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting blocked user: %v", err))
		return &common.InternalServerError{Message: "Error getting blocked user"}

	}
	if blockedUser == nil {
		slog.Error(fmt.Sprintf("Blocked user %s not found", req.BlockedUserId))
		return &common.NotFoundError{Message: "Blocked user not found"}
	}

	// check if already blocked
	if !user.BlockedUsers[req.BlockedUserId] {
		slog.Error(fmt.Sprintf("User %s is not blocked", req.BlockedUserId))
		return &common.BadRequestError{Message: "User is not blocked"}
	}

	// unblock the user in the database
	err = dbClient.UnBlockUser(ctx, *user, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error unblocking user: %v", err))
		return &common.InternalServerError{Message: "Error unblocking user"}
	}
	return nil
}

func blockUser(ctx context.Context, userId string, req BlockUserRequest) error {

	user, err := dbClient.GetUser(ctx, userId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", userId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// get blocked user
	blockedUser, err := dbClient.GetUser(ctx, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting blocked user: %v", err))
		return &common.InternalServerError{Message: "Error getting blocked user"}

	}
	if blockedUser == nil {
		slog.Error(fmt.Sprintf("Blocked user %s not found", req.BlockedUserId))
		return &common.NotFoundError{Message: "Blocked user not found"}
	}

	// check if already blocked
	if user.BlockedUsers[req.BlockedUserId] {
		slog.Error(fmt.Sprintf("User %s is already blocked", req.BlockedUserId))
		return &common.BadRequestError{Message: "User is already blocked"}
	}

	// block the user in the database
	err = dbClient.BlockUser(ctx, *user, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error blocking user: %v", err))
		return &common.InternalServerError{Message: "Error blocking user"}
	}
	return nil
}
