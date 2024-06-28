package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
POST /v1/groups/create
*/
func createGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var group createGroupRequest
	err := decoder.Decode(&group)
	if err != nil || group.GroupName == "" {
		logger.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	groupId := fmt.Sprintf("group-%s", uuid.New().String())

	dbGroup := dbGroup{
		GroupId:   groupId,
		GroupName: group.GroupName,
		Members:   map[string]bool{},
	}

	err = dbClient.StoreGroup(r.Context(), dbGroup)
	if err != nil {
		// log error
		logger.Error(fmt.Sprintf("Error storing group: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Infof("Group created: %v", groupId)

	// return the group ID and name in the response
	resp := createGroupResponse{
		GroupId:   groupId,
		GroupName: group.GroupName,
	}
	json.NewEncoder(w).Encode(resp)

}

type userToGroupRequest struct {
	UserId string `json:"UserId"`
}

/*
POST /v1/groups/:groupId/add
*/
func addUserToGroup(w http.ResponseWriter, r *http.Request) {
	// get the group ID from the URL path
	groupId := chi.URLParam(r, "groupId")

	if groupId == "" {
		logger.Error("Group ID is required")
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req userToGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserId == "" {
		logger.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	group, err := dbClient.GetGroup(r.Context(), groupId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if group == nil {
		logger.Error(fmt.Sprintf("Group %s not found", groupId))
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	user, err := dbClient.GetUser(r.Context(), req.UserId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		logger.Error(fmt.Sprintf("User %s not found", req.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// check if user is already a member
	if group.Members[req.UserId] {
		logger.Error(fmt.Sprintf("User %d is already a member of the group %d", req.UserId, groupId))
		http.Error(w, "User is already a member of the group", http.StatusBadRequest)
		return
	}

	// add user to group
	err = dbClient.AddUserToGroup(r.Context(), *group, *user)
	if err != nil {
		logger.Error(fmt.Sprintf("Error adding %s user to group %s : %v", req.UserId, groupId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Infof("User %s added to group: %s", req.UserId, groupId)

	w.WriteHeader(http.StatusOK)
}

/*
POST /v1/groups/:groupId/remove
*/
func removeUserFromGroup(w http.ResponseWriter, r *http.Request) {
	// get the group ID from the URL path
	groupId := chi.URLParam(r, "groupId")

	if groupId == "" {
		logger.Error("Group ID is required")
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req userToGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserId == "" {
		logger.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	group, err := dbClient.GetGroup(r.Context(), groupId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if group == nil {
		logger.Error(fmt.Sprintf("Group %s not found", groupId))
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	user, err := dbClient.GetUser(r.Context(), req.UserId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		logger.Error(fmt.Sprintf("User %s not found", req.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// check if user is not a member
	if !group.Members[req.UserId] {
		logger.Error(fmt.Sprintf("User %d is not a member of the group %d", req.UserId, groupId))
		http.Error(w, "User is not a member of the group", http.StatusBadRequest)
		return
	}

	// remove user from group
	err = dbClient.RemoveUserFromGroup(r.Context(), *group, *user)
	if err != nil {
		logger.Error(fmt.Sprintf("Error removing %s user from group %s : %v", req.UserId, groupId, err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Infof("User %s removed from group: %s", req.UserId, groupId)

}
