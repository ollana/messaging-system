package messages

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	. "server/common"
	"server/db"
	"time"
)

type SendMessageRequest struct {
	SenderId    string `json:"senderId"`
	RecipientId string `json:"recipientId"`
	Message     string `json:"message"`
}

type HandlerInterface interface {
	SendPrivateMessage(ctx context.Context, req SendMessageRequest) error
	SendGroupMessage(ctx context.Context, req SendMessageRequest) error
	GetMessages(ctx context.Context, recipientId string, timestamp int64) (*UserMessagesResp, error)
}

type Handler struct {
	DBClient db.DynamoDBClientInterface
}

/*
Send a private message to a user
If the recipient has blocked the sender, return 403 Forbidden
*/
func (handler *Handler) SendPrivateMessage(ctx context.Context, req SendMessageRequest) error {

	recipient, err := handler.DBClient.GetUser(ctx, req.RecipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting recipient user: %v", err))
		return &InternalServerError{Message: "Error getting recipient user"}
	}
	if recipient == nil {
		slog.Error(fmt.Sprintf("Recipient user not found: %v", req.RecipientId))
		return &NotFoundError{Message: "Recipient not found"}
	}
	// check if the recipient has blocked the sender
	if recipient.BlockedUsers[req.SenderId] {
		slog.Error(fmt.Sprintf("Recipient %s has blocked sender %s", req.RecipientId, req.SenderId))
		return &ForbiddenError{Message: "Recipient has blocked the sender"}
	}

	// validate sender and recipient exists
	sender, err := handler.DBClient.GetUser(ctx, req.SenderId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting sender: %v", err))
		return &InternalServerError{Message: "Error getting sender"}
	}
	if sender == nil {
		slog.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		return &NotFoundError{Message: "Sender not found"}
	}

	msg := Message{
		RecipientId: req.RecipientId,
		Timestamp:   time.Now().Format(time.RFC3339), // store the dates in RFC339 string format so that they can be both human-readable and easy to query.
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = handler.DBClient.StoreMessage(ctx, msg)
	if err != nil {
		slog.Error(fmt.Sprintf("Error storing message: %v", err))
		return &InternalServerError{Message: "Error storing message"}
	}

	slog.Info(fmt.Sprintf("Message sent from %s to user %s", req.SenderId, req.RecipientId))

	return nil
}

/*
Send a group message
If the sender is not a member of the group, return 403 Forbidden
*/
func (handler *Handler) SendGroupMessage(ctx context.Context, req SendMessageRequest) error {
	// validate sender and recipient exists
	sender, err := handler.DBClient.GetUser(ctx, req.SenderId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting sender: %v", err))
		return &InternalServerError{Message: "Error getting sender"}
	}
	if sender == nil {
		slog.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		return &NotFoundError{Message: "Sender not found"}
	}

	recipient, err := handler.DBClient.GetGroup(ctx, req.RecipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting recipient group: %v", err))
		return &InternalServerError{Message: "Error getting recipient group"}
	}
	if recipient == nil {
		slog.Error(fmt.Sprintf("Recipient group not found: %v", req.RecipientId))
		return &NotFoundError{Message: "Recipient not found"}
	}

	// check if the sender is a member of the group
	if !sender.Groups[req.RecipientId] {
		slog.Error(fmt.Sprintf("Sender %s is not a member of group %s", req.SenderId, req.RecipientId))
		return &ForbiddenError{Message: "Sender is not a member of the group"}
	}

	msg := Message{
		RecipientId: req.RecipientId,
		Timestamp:   time.Now().Format(time.RFC3339),
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = handler.DBClient.StoreMessage(ctx, msg)
	if err != nil {
		slog.Error(fmt.Sprintf("Error storing message: %v", err))
		return &InternalServerError{Message: "Error storing message"}
	}

	slog.Info(fmt.Sprintf("Message sent from %s to group %s", req.SenderId, req.RecipientId))
	return nil
}

type UserMessagesResp struct {
	Messages []Message `json:"messages"`
}

func (handler *Handler) GetMessages(ctx context.Context, recipientId string, timestamp int64) (*UserMessagesResp, error) {

	user, err := handler.DBClient.GetUser(ctx, recipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return nil, &InternalServerError{Message: "Error getting user"}

	}
	if user == nil {
		slog.Error(fmt.Sprintf("User not found: %v", recipientId))
		return nil, &NotFoundError{Message: "User not found"}
	}

	messages, err := handler.DBClient.GetMessages(ctx, *user, timestamp)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting messages: %v", err))
		return nil, &InternalServerError{Message: "Error getting messages"}
	}

	resp := UserMessagesResp{
		Messages: messages,
	}

	slog.Info(fmt.Sprintf("Total of %d messages retrieved for user %s", len(messages), recipientId))

	return &resp, nil
}
