package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"net/http"
	"server/common"
	"server/users"
)

type UsersRoutes struct {
	Handler users.UsersHandlerInterface
}

/*
Create a new user
API: POST /v1/users/create
*/
func (ur *UsersRoutes) CreateUserHandler(c *gin.Context) {
	// read the request body
	decoder := json.NewDecoder(c.Request.Body)
	var req users.RegisterUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.UserName == "" {
		slog.Error(fmt.Sprintf("Invalid input %v", req))
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	resp, err := ur.Handler.RegisterUser(c, req)
	if err != nil {
		common.HandleError(err, c)
		return
	}
	slog.Info(fmt.Sprintf("User created: %s", resp.UserId))
	c.JSON(http.StatusOK, resp)
}

/*
Block a user for the given user ID, op can be block or unblock
API: POST /v1/users/:userId?op=[block/unblock]
*/
func (ur *UsersRoutes) BlockUserHandler(c *gin.Context) {
	// get the user ID from the URL path
	userId := c.Param("userId")
	if userId == "" {
		c.String(http.StatusBadRequest, "userId is required")
		return
	}

	// read the request body
	decoder := json.NewDecoder(c.Request.Body)
	var req users.BlockUserRequest
	err := decoder.Decode(&req)
	if err != nil || req.BlockedUserId == "" {
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	op := c.Query("op")

	switch op {
	case "block":
		err = ur.Handler.BlockUser(c, userId, req)
	case "unblock":
		err = ur.Handler.UnblockUser(c, userId, req)
	default:
		slog.Error(fmt.Sprintf("Invalid operation: %v", op))
		c.String(http.StatusBadRequest, "Invalid operation")
		return
	}

	if err != nil {
		common.HandleError(err, c)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
