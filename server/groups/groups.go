package groups

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"server/common"
	"server/db"
)

type GroupHandlerInterface interface {
	CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error)
	AddUserToGroup(ctx context.Context, groupId string, req *UserToGroupRequest) error
	RemoveUserFromGroup(ctx context.Context, groupId string, req *UserToGroupRequest) error
}

type GroupHandler struct {
	DBClient db.DynamoDBClientInterface
}

type UserToGroupRequest struct {
	UserId string `json:"userId"`
}

type CreateGroupRequest struct {
	GroupName string `json:"groupName"`
}

type CreateGroupResponse struct {
	GroupId   string `json:"groupId"`
	GroupName string `json:"groupName"`
}

func (handler *GroupHandler) CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error) {

	groupId := fmt.Sprintf("group-%s", uuid.New().String())

	dbGroup := db.Group{
		GroupId:   groupId,
		GroupName: req.GroupName,
		Members:   map[string]bool{},
	}

	err := handler.DBClient.StoreGroup(ctx, dbGroup)
	if err != nil {
		// log error
		slog.Error(fmt.Sprintf("Error storing group: %v", err))
		return nil, &common.InternalServerError{Message: "Error storing group"}
	}
	slog.Info(fmt.Sprintf("Group created: %v", groupId))

	// return the group ID and name in the response
	resp := CreateGroupResponse{
		GroupId:   groupId,
		GroupName: req.GroupName,
	}
	return &resp, nil
}

func (handler *GroupHandler) AddUserToGroup(ctx context.Context, groupId string, req *UserToGroupRequest) error {

	group, err := handler.DBClient.GetGroup(ctx, groupId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		return &common.InternalServerError{Message: "Error getting group"}
	}
	if group == nil {
		slog.Error(fmt.Sprintf("Group %s not found", groupId))
		return &common.NotFoundError{Message: "Group not found"}
	}

	user, err := handler.DBClient.GetUser(ctx, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", req.UserId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// check if user is already a member
	if group.Members[req.UserId] {
		slog.Error(fmt.Sprintf("User %s is already a member of the group %s", req.UserId, groupId))
		return &common.BadRequestError{Message: "User is already a member of the group"}
	}

	// add user to group
	err = handler.DBClient.AddUserToGroup(ctx, *group, *user)
	if err != nil {
		slog.Error(fmt.Sprintf("Error adding %s user to group %s : %v", req.UserId, groupId, err))
		return &common.InternalServerError{Message: "Error adding user to group"}
	}
	slog.Info(fmt.Sprintf("User %s added to group: %s", req.UserId, groupId))

	return nil
}

func (handler *GroupHandler) RemoveUserFromGroup(ctx context.Context, groupId string, req *UserToGroupRequest) error {
	group, err := handler.DBClient.GetGroup(ctx, groupId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting group %s : %v", groupId, err))
		return &common.InternalServerError{Message: "Error getting group"}

	}
	if group == nil {
		slog.Error(fmt.Sprintf("Group %s not found", groupId))
		return &common.NotFoundError{Message: "Group not found"}
	}

	user, err := handler.DBClient.GetUser(ctx, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user %s: %v", req.UserId, err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", req.UserId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// check if user is not a member
	if !group.Members[req.UserId] {
		slog.Error(fmt.Sprintf("User %s is not a member of the group %s", req.UserId, groupId))
		return &common.BadRequestError{Message: "User is not a member of the group"}
	}

	// remove user from group
	err = handler.DBClient.RemoveUserFromGroup(ctx, *group, *user)
	if err != nil {
		slog.Error(fmt.Sprintf("Error removing %s user from group %s : %v", req.UserId, groupId, err))
		return &common.InternalServerError{Message: "Error removing user from group"}
	}
	slog.Info(fmt.Sprintf("User %s removed from group: %s", req.UserId, groupId))

	return nil
}
