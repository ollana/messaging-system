package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type sendMessageRequest struct {
	SenderId    string `json:"SenderId"`
	RecipientId string `json:"RecipientId"`
	Message     string `json:"Message"`
}

/*
Send a private message to a user
If blocked by the recipient, return 403 Forbidden
API: POST /v1/messages/private
*/
func sendPrivateMessage(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var req sendMessageRequest
	err := decoder.Decode(&req)
	if err != nil || req.SenderId == "" || req.RecipientId == "" || req.Message == "" {
		logger.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	recipient, err := dbClient.GetUser(r.Context(), req.RecipientId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting recipient user: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if recipient == nil {
		logger.Error(fmt.Sprintf("Recipient user not found: %v", req.RecipientId))
		http.Error(w, "Recipient not found", http.StatusBadRequest)
		return
	}
	// check if the recipient has blocked the sender
	if recipient.BlockedUsers[req.SenderId] {
		logger.Error(fmt.Sprintf("Recipient %s has blocked sender %s", req.RecipientId, req.SenderId))
		http.Error(w, "Recipient has blocked the sender", http.StatusForbidden)
		return
	}

	// validate sender and recipient exists
	sender, err := dbClient.GetUser(r.Context(), req.SenderId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting sender: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if sender == nil {
		logger.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		http.Error(w, "Sender not found", http.StatusBadRequest)
		return
	}

	msg := dbMessage{
		RecipientId: req.RecipientId,
		Timestamp:   fmt.Sprintf("%d", time.Now().Unix()),
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = dbClient.StoreMessage(r.Context(), msg)
	if err != nil {
		logger.Error(fmt.Sprintf("Error storing message: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Infof("Message sent from %s to user %s", req.SenderId, req.RecipientId)

	w.WriteHeader(http.StatusOK)
}

/*
Send a message to a group
If the sender is not a member of the group, return 403 Forbidden
API: POST /v1/messages/group
*/
func sendGroupMessage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req sendMessageRequest
	err := decoder.Decode(&req)
	if err != nil || req.SenderId == "" || req.RecipientId == "" || req.Message == "" {
		logger.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// validate sender and recipient exists
	sender, err := dbClient.GetUser(r.Context(), req.SenderId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting sender: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if sender == nil {
		logger.Error(fmt.Sprintf("Sender not found: %v", req.SenderId))
		http.Error(w, "Sender not found", http.StatusBadRequest)
		return
	}

	// check if the sender is a member of the group
	if !sender.Groups[req.RecipientId] {
		logger.Error(fmt.Sprintf("Sender %s is not a member of group %s", req.SenderId, req.RecipientId))
		http.Error(w, "Sender is not a member of the group", http.StatusForbidden)
		return
	}

	recipient, err := dbClient.GetGroup(r.Context(), req.RecipientId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting recipient group: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if recipient == nil {
		logger.Error(fmt.Sprintf("Recipient group not found: %v", req.RecipientId))
		http.Error(w, "Recipient not found", http.StatusBadRequest)
		return
	}

	msg := dbMessage{
		RecipientId: req.RecipientId,
		Timestamp:   fmt.Sprintf("%d", time.Now().Unix()),
		SenderId:    req.SenderId,
		Message:     req.Message,
	}

	err = dbClient.StoreMessage(r.Context(), msg)
	if err != nil {
		logger.Error(fmt.Sprintf("Error storing message: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Infof("Message sent from %s to group %s", req.SenderId, req.RecipientId)

	w.WriteHeader(http.StatusOK)
}

type userMessages struct {
	Messages []dbMessage `json:"Messages"`
}

/*
Get all messages for a user by userId, including private messages and group messages
API: GET /v1/messages/:userId
*/
func getMessages(w http.ResponseWriter, r *http.Request) {
	recipientId := chi.URLParam(r, "userId")
	if recipientId == "" {
		logger.Error("userId is required")
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	user, err := dbClient.GetUser(r.Context(), recipientId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting user: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		logger.Error(fmt.Sprintf("User not found: %v", recipientId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	messages, err := dbClient.GetMessages(r.Context(), *user)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting messages: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := userMessages{
		Messages: messages,
	}

	json.NewEncoder(w).Encode(resp)
}
