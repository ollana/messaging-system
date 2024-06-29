package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
API: POST /v1/messages/send?type=[private/group]
*/
func (mr *MessagesRoutes) SendMessageHandler(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var req messages.SendMessageRequest
	err := decoder.Decode(&req)
	if err != nil || req.SenderId == "" || req.RecipientId == "" || req.Message == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", c.Request.Body))
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	msgType := c.Query("type")
	switch msgType {
	case "private":
		err = mr.Handler.SendPrivateMessage(c, req)
	case "group":
		err = mr.Handler.SendGroupMessage(c, req)
	default:
		slog.Error(fmt.Sprintf("Invalid type %s", msgType))
		c.String(http.StatusBadRequest, "Invalid operation")
		return
	}
	if err != nil {
		common.HandleError(err, c)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}

/*
Get all messages for a user by userId, including private messages and group messages
Optional query parameter timestamp, to get messages after a certain timestamp
API: GET /v1/messages/:userId?timestamp=123456
*/
func (mr *MessagesRoutes) GetMessagesHandler(c *gin.Context) {
	recipientId := c.Param("userId")
	timestamp := c.Query("timestamp")
	if recipientId == "" {
		slog.Error("userId is required")
		c.String(http.StatusBadRequest, "userId is required")
		return
	}
	var unixTimeStemp int64
	if timestamp != "" {
		// parse timestamp to int64 and validate
		i, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid timestamp: %v", timestamp))
			c.String(http.StatusBadRequest, "Invalid timestamp")
			return
		}
		unixTimeStemp = i
	}

	resp, err := mr.Handler.GetMessages(c, recipientId, unixTimeStemp)
	if err != nil {
		common.HandleError(err, c)
		return
	}
	c.JSON(http.StatusOK, resp)
}
