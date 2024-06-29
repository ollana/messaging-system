package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/db"
	"strconv"
	"time"
)

type sendMessageRequest struct {
	SenderId    string `json:"SenderId"`
	RecipientId string `json:"RecipientId"`
	Message     string `json:"Message"`
}

/*
Send a private or group message, type can be [group/private]
API: POST /v1/messages/{type}
*/
func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req sendMessageRequest
	err := decoder.Decode(&req)
	if err != nil || req.SenderId == "" || req.RecipientId == "" || req.Message == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	msgType := chi.URLParam(r, "type")
	switch msgType {
	case "private":
		err = sendPrivateMessage(r.Context(), req)
	case "group":
		err = sendGroupMessage(r.Context(), req)
	default:
		slog.Error(fmt.Sprintf("Invalid type %s", msgType))
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}
	if err != nil {
		common.HandleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
Send a private message to a user
If the recipient has blocked the sender, return 403 Forbidden
*/
func sendPrivateMessage(ctx context.Context, req sendMessageRequest) error {

	recipient, err := dbClient.GetUser(ctx, req.RecipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting recipient user: %v", err))
		return &common.InternalServerError{Message: "Error getting recipient user"}
	}
	if recipient == nil {
		slog.Error(fmt.Sprintf("Recipient user not found: %v", req.RecipientId))
		return &common.NotFoundError{Message: "Recipient not found"}
	}
	// check if the recipient has blocked the sender
	if recipient.BlockedUsers[req.SenderId] {
		slog.Error(fmt.Sprintf("Recipient %s has blocked sender %s", req.RecipientId, req.SenderId))
		return &common.ForbiddenError{Message: "Recipient has blocked the sender"}
	}

	// validate sender and recipient exists
	sender, err := dbClient.GetUser(ctx, req.SenderId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting sender: %v", err))
		return &common.InternalServerError{Message: "Error getting sender"}
	}
	if sender == nil {
		slog.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		return &common.NotFoundError{Message: "Sender not found"}
	}

	msg := db.Message{
		RecipientId: req.RecipientId,
		Timestamp:   time.Now().Format(time.RFC3339), // store the dates in RFC339 string format so that they can be both human-readable and easy to query.
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = dbClient.StoreMessage(ctx, msg)
	if err != nil {
		slog.Error(fmt.Sprintf("Error storing message: %v", err))
		return &common.InternalServerError{Message: "Error storing message"}
	}

	slog.Info("Message sent from %s to user %s", req.SenderId, req.RecipientId)

	return nil
}

/*
Send a group message
If the sender is not a member of the group, return 403 Forbidden
*/
func sendGroupMessage(ctx context.Context, req sendMessageRequest) error {
	// validate sender and recipient exists
	sender, err := dbClient.GetUser(ctx, req.SenderId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting sender: %v", err))
		return &common.InternalServerError{Message: "Error getting sender"}
	}
	if sender == nil {
		slog.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		return &common.NotFoundError{Message: "Sender not found"}
	}

	recipient, err := dbClient.GetGroup(ctx, req.RecipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting recipient group: %v", err))
		return &common.InternalServerError{Message: "Error getting recipient group"}
	}
	if recipient == nil {
		slog.Error(fmt.Sprintf("Recipient group not found: %v", req.RecipientId))
		return &common.NotFoundError{Message: "Recipient not found"}
	}

	// check if the sender is a member of the group
	if !sender.Groups[req.RecipientId] {
		slog.Error(fmt.Sprintf("Sender %s is not a member of group %s", req.SenderId, req.RecipientId))
		return &common.ForbiddenError{Message: "Sender is not a member of the group"}
	}

	msg := db.Message{
		RecipientId: req.RecipientId,
		Timestamp:   fmt.Sprintf("%d", time.Now().Unix()),
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = dbClient.StoreMessage(ctx, msg)
	if err != nil {
		slog.Error(fmt.Sprintf("Error storing message: %v", err))
		return &common.InternalServerError{Message: "Error storing message"}
	}

	slog.Info("Message sent from %s to group %s", req.SenderId, req.RecipientId)
	return nil
}

type userMessages struct {
	Messages []db.Message `json:"Messages"`
}

/*
Get all messages for a user by userId, including private messages and group messages
Optional query parameter timestamp, to get messages after a certain timestamp
API: GET /v1/messages/:userId?timestamp=123456
*/
func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	recipientId := chi.URLParam(r, "userId")
	timestamp := r.URL.Query().Get("timestamp")
	if recipientId == "" {
		slog.Error("userId is required")
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}
	var unixTimeStemp int64
	if timestamp != "" {
		// parse timestamp to int64 and validate
		i, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid timestamp: %v", timestamp))
			http.Error(w, "Invalid timestamp", http.StatusBadRequest)
			return
		}
		unixTimeStemp = i
	}

	resp, err := getMessages(r.Context(), recipientId, unixTimeStemp)
	if err != nil {
		common.HandleError(err, w)
		return
	}
	json.NewEncoder(w).Encode(resp)
}
func getMessages(ctx context.Context, recipientId string, timestamp int64) (*userMessages, error) {

	user, err := dbClient.GetUser(ctx, recipientId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v", err))
		return nil, &common.InternalServerError{Message: "Error getting user"}

	}
	if user == nil {
		slog.Error(fmt.Sprintf("User not found: %v", recipientId))
		return nil, &common.NotFoundError{Message: "User not found"}
	}

	messages, err := dbClient.GetMessages(ctx, *user, timestamp)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting messages: %v", err))
		return nil, &common.InternalServerError{Message: "Error getting messages"}
	}

	resp := userMessages{
		Messages: messages,
	}

	return &resp, nil
}
