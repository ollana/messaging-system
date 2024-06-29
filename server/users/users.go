package users

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"server/common"
	"server/db"
)

type RegisterUserRequest struct {
	UserName string `json:"userName"`
}
type RegisterUserResponse struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
}

type BlockUserRequest struct {
	BlockedUserId string `json:"blockedUserId"`
}

type UsersHandlerInterface interface {
	RegisterUser(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error)
	BlockUser(ctx context.Context, userId string, req BlockUserRequest) error
	UnblockUser(ctx context.Context, userId string, req BlockUserRequest) error
}

type UsersHandler struct {
	DBClient db.DynamoDBClientInterface
}

func (handler *UsersHandler) RegisterUser(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error) {
	// generate a new user ID with UUID
	userId := fmt.Sprintf("user-%s", uuid.New().String())

	// store the user in the database
	user := db.User{
		UserId:       userId,
		UserName:     req.UserName,
		BlockedUsers: make(map[string]bool),
	}
	err := handler.DBClient.StoreUser(ctx, user)
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

func (handler *UsersHandler) UnblockUser(ctx context.Context, userId string, req BlockUserRequest) error {

	user, err := handler.DBClient.GetUser(ctx, userId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", userId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// get blocked user
	blockedUser, err := handler.DBClient.GetUser(ctx, req.BlockedUserId)
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
	err = handler.DBClient.UnBlockUser(ctx, *user, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error unblocking user: %v", err))
		return &common.InternalServerError{Message: "Error unblocking user"}
	}
	slog.Info(fmt.Sprintf("User %s unblocked for %s", req.BlockedUserId, userId))
	return nil
}

func (handler *UsersHandler) BlockUser(ctx context.Context, userId string, req BlockUserRequest) error {

	user, err := handler.DBClient.GetUser(ctx, userId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return &common.InternalServerError{Message: "Error getting user"}
	}
	if user == nil {
		slog.Error(fmt.Sprintf("User %s not found", userId))
		return &common.NotFoundError{Message: "User not found"}
	}

	// get blocked user
	blockedUser, err := handler.DBClient.GetUser(ctx, req.BlockedUserId)
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
	err = handler.DBClient.BlockUser(ctx, *user, req.BlockedUserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error blocking user: %v", err))
		return &common.InternalServerError{Message: "Error blocking user"}
	}

	slog.Info(fmt.Sprintf("User %s blocked for %s", req.BlockedUserId, userId))

	return nil
}
