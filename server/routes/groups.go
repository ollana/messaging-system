package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
func (gr GroupRoutes) CreateGroupHandler(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var req groups.CreateGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.GroupName == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", c.Request.Body))
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	resp, err := gr.Handler.CreateGroup(c, &req)
	if err != nil {
		common.HandleError(err, c)
		return
	}
	c.JSON(http.StatusOK, resp)
}

/*
Add a user to a group or remove a user from a group
API: POST /v1/groups/:groupId?op=[add/remove]
*/
func (gr GroupRoutes) UserToGroupHandler(c *gin.Context) {
	// get the group ID from the URL path
	groupId := c.Param("groupId")

	if groupId == "" {
		slog.Error("Group ID is required")
		c.String(http.StatusBadRequest, "Group ID is required")
		return
	}

	// read the request body
	decoder := json.NewDecoder(c.Request.Body)
	var req groups.UserToGroupRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserId == "" {
		slog.Error(fmt.Sprintf("Invalid input: %v", c.Request.Body))
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}

	// switch on the URL path to determine the operation
	op := c.Query("op")
	switch op {
	case "add":
		err = gr.Handler.AddUserToGroup(c, groupId, &req)
	case "remove":
		err = gr.Handler.RemoveUserFromGroup(c, groupId, &req)
	default:
		slog.Error(fmt.Sprintf("Invalid operation %s", op))
		c.String(http.StatusBadRequest, "Invalid operation")
		return
	}
	if err != nil {
		common.HandleError(err, c)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
