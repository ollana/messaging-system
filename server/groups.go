package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"net/http"
)

type createGroupRequest struct {
	GroupName string `json:"GroupName"`
}

type createGroupResponse struct {
	GroupId   string `json:"GroupId"`
	GroupName string `json:"GroupName"`
}

/*
Create a new group
API: POST /v1/groups/create
*/

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req createGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.GroupName == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	resp, err := createGroup(r.Context(), &req)
	if err != nil {
		handleError(err, w)
		return
	}
	json.NewEncoder(w).Encode(resp)

}
func createGroup(ctx context.Context, req *createGroupRequest) (*createGroupResponse, error) {

	groupId := fmt.Sprintf("group-%s", uuid.New().String())

	dbGroup := dbGroup{
		GroupId:   groupId,
		GroupName: req.GroupName,
		Members:   map[string]bool{},
	}

	err := dbClient.StoreGroup(ctx, dbGroup)
	if err != nil {
		// log error
		slog.Error(fmt.Sprintf("Error storing group: %v", err))
		return nil, &InternalServerError{Message: "Error storing group"}
	}
	slog.Info("Group created: %v", groupId)

	// return the group ID and name in the response
	resp := createGroupResponse{
		GroupId:   groupId,
		GroupName: req.GroupName,
	}
	return &resp, nil
}

type userToGroupRequest struct {
	UserId string `json:"UserId"`
}

/*
Add a user to a group or remove a user from a group
API: POST /v1/groups/:groupId/[add/remove]
*/
func addUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	// get the group ID from the URL path
	groupId := chi.URLParam(r, "groupId")

	if groupId == "" {
		slog.Error("Group ID is required")
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req userToGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserId == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// switch on the URL path to determine the operation
	op := chi.URLParam(r, "op")
	switch chi.URLParam(r, "op") {
	case "add":
		err = addUserToGroup(r.Context(), groupId, &req)
	case "remove":
		err = removeUserFromGroup(r.Context(), groupId, &req)
	default:
		slog.Error(fmt.Sprintf("Invalid operation %s", op))
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}
	if err != nil {
		handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)

	w.WriteHeader(http.StatusOK)
}

func addUserToGroup(ctx context.Context, groupId string, req *userToGroupRequest) error {

	group, err := dbClient.GetGroup(ctx, groupId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		return &InternalServerError{Message: "Error getting group"}
	}
	if group == nil {
		slog.Error(fmt.Sprintf("Group %s not found", groupId))
		return &NotFoundError{Message: "Group not found"}
	}

	user, err := dbClient.GetUser(ctx, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		return &InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", req.UserId))
		return &NotFoundError{Message: "User not found"}
	}

	// check if user is already a member
	if group.Members[req.UserId] {
		slog.Error(fmt.Sprintf("User %s is already a member of the group %s", req.UserId, groupId))
		return &BadRequestError{Message: "User is already a member of the group"}
	}

	// add user to group
	err = dbClient.AddUserToGroup(ctx, *group, *user)
	if err != nil {
		slog.Error(fmt.Sprintf("Error adding %s user to group %s : %v", req.UserId, groupId, err))
		return &InternalServerError{Message: "Error adding user to group"}
	}
	slog.Info("User %s added to group: %s", req.UserId, groupId)

	return nil
}

func removeUserFromGroup(ctx context.Context, groupId string, req *userToGroupRequest) error {
	group, err := dbClient.GetGroup(ctx, groupId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		return &InternalServerError{Message: "Error getting group"}

	}
	if group == nil {
		slog.Error(fmt.Sprintf("Group %s not found", groupId))
		return &NotFoundError{Message: "Group not found"}
	}

	user, err := dbClient.GetUser(ctx, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		return &InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", req.UserId))
		return &NotFoundError{Message: "User not found"}
	}

	// check if user is not a member
	if !group.Members[req.UserId] {
		slog.Error(fmt.Sprintf("User %s is not a member of the group %s", req.UserId, groupId))
		return &BadRequestError{Message: "User is not a member of the group"}
	}

	// remove user from group
	err = dbClient.RemoveUserFromGroup(ctx, *group, *user)
	if err != nil {
		slog.Error(fmt.Sprintf("Error removing %s user from group %s : %v", req.UserId, groupId, err))
		return &InternalServerError{Message: "Error removing user from group"}
	}
	slog.Info("User %s removed from group: %s", req.UserId, groupId)

	return nil
}
