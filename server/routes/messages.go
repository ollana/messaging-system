package routes

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/messages"
	"strconv"
)

type MessagesRoutes struct {
	Handler messages.HandlerInterface
}

/*
Send a private or group message, type can be [group/private]
API: POST /v1/messages/{type}
*/
func (mr *MessagesRoutes) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req messages.SendMessageRequest
	err := decoder.Decode(&req)
	if err != nil || req.SenderId == "" || req.RecipientId == "" || req.Message == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", r.Body))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	msgType := chi.URLParam(r, "type")
	switch msgType {
	case "private":
		err = mr.Handler.SendPrivateMessage(r.Context(), req)
	case "group":
		err = mr.Handler.SendGroupMessage(r.Context(), req)
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
Get all messages for a user by userId, including private messages and group messages
Optional query parameter timestamp, to get messages after a certain timestamp
API: GET /v1/messages/:userId?timestamp=123456
*/
func (mr *MessagesRoutes) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
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

	resp, err := mr.Handler.GetMessages(r.Context(), recipientId, unixTimeStemp)
	if err != nil {
		common.HandleError(err, w)
		return
	}
	json.NewEncoder(w).Encode(resp)
}
