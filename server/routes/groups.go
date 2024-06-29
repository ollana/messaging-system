package routes

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/groups"
)

type GroupRoutes struct {
	Handler groups.GroupHandlerInterface
}

/*
Create a new group
API: POST /v1/groups/create
*/
func (gr GroupRoutes) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req groups.CreateGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.GroupName == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	resp, err := gr.Handler.CreateGroup(r.Context(), &req)
	if err != nil {
		common.HandleError(err, w)
		return
	}
	json.NewEncoder(w).Encode(resp)

}

/*
Add a user to a group or remove a user from a group
API: POST /v1/groups/:groupId/[add/remove]
*/
func (gr GroupRoutes) UserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	// get the group ID from the URL path
	groupId := chi.URLParam(r, "groupId")

	if groupId == "" {
		slog.Error("Group ID is required")
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// read the request body
	decoder := json.NewDecoder(r.Body)
	var req groups.UserToGroupRequest
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
		err = gr.Handler.AddUserToGroup(r.Context(), groupId, &req)
	case "remove":
		err = gr.Handler.RemoveUserFromGroup(r.Context(), groupId, &req)
	default:
		slog.Error(fmt.Sprintf("Invalid operation %s", op))
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}
	if err != nil {
		common.HandleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)

	w.WriteHeader(http.StatusOK)
}
