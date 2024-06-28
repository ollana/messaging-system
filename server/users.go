package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
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
func registerUser(w http.ResponseWriter, r *http.Request) {
	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req RegisterUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserName == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// generate a new user ID with UUID
	userId := fmt.Sprintf("user-%s", uuid.New().String())

	// store the user in the database
	user := dbUser{
		UserId:       userId,
		UserName:     req.UserName,
		BlockedUsers: make(map[string]bool),
	}
	err = dbClient.StoreUser(r.Context(), user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// return the user ID and name in the response
	resp := RegisterUserResponse{
		UserId:   userId,
		UserName: req.UserName,
	}
	json.NewEncoder(w).Encode(resp)
}

type BlockUserRequest struct {
	BlockedUserId string `json:"BlockedUserId"`
}

/*
Block a user for the given user ID
API: POST /v1/users/:userId/block
*/
func blockUser(w http.ResponseWriter, r *http.Request) {

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
	user, err := dbClient.GetUser(r.Context(), userId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// get blocked user
	blockedUser, err := dbClient.GetUser(r.Context(), req.BlockedUserId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if blockedUser == nil {
		http.Error(w, "Blocked user not found", http.StatusNotFound)
		return
	}

	// check if already blocked
	if user.BlockedUsers[req.BlockedUserId] {
		http.Error(w, "User is already blocked", http.StatusBadRequest)
		return
	}

	// block the user in the database
	err = dbClient.BlockUser(r.Context(), *user, req.BlockedUserId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
