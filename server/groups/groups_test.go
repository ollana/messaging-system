package groups

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"server/common"
	. "server/common"
	"server/db"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	ctx := context.Background()
	handler := GroupHandler{
		DBClient: db.NewMockDBClient(),
	}
	t.Run("Create group successfully", func(t *testing.T) {
		// create a new request
		req := CreateGroupRequest{
			GroupName: "test-group",
		}

		// seed the request to the handler
		resp, err := handler.CreateGroup(ctx, &req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.GroupId)
		assert.Equal(t, req.GroupName, resp.GroupName)

	})

	t.Run("db error", func(t *testing.T) {

		handler := GroupHandler{
			DBClient: db.NewMockDBClient(),
		}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := CreateGroupRequest{
			GroupName: "test-group",
		}
		resp, err := handler.CreateGroup(ctx, &req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)
		assert.Nil(t, resp)

	})

}

func TestAddUserToGroup(t *testing.T) {
	ctx := context.Background()
	handler := GroupHandler{
		DBClient: db.NewMockDBClient(),
	}
	t.Run("Add user to group successfully", func(t *testing.T) {
		handler.DBClient.StoreUser(context.Background(), User{UserId: "test-user-1"})
		handler.DBClient.StoreGroup(context.Background(), Group{GroupId: "test-group-1"})
		// create a new request
		req := UserToGroupRequest{
			UserId: "test-user-1",
		}

		err := handler.AddUserToGroup(ctx, "test-group-1", &req)
		assert.NoError(t, err)

		user, _ := handler.DBClient.GetUser(ctx, "test-user-1")
		assert.Contains(t, user.Groups, "test-group-1")
		group, _ := handler.DBClient.GetGroup(ctx, "test-group-1")
		assert.Contains(t, group.Members, "test-user-1")

	})

	t.Run("invalid group", func(t *testing.T) {
		// create a new request
		handler.DBClient.StoreUser(ctx, User{UserId: "test-user-2"})
		req := UserToGroupRequest{
			UserId: "test-user-2",
		}
		err := handler.AddUserToGroup(ctx, "test-group-2", &req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("invalid user", func(t *testing.T) {
		handler.DBClient.StoreGroup(ctx, Group{GroupId: "test-group-3"})
		// create a new request
		req := UserToGroupRequest{
			UserId: "test-user-3",
		}
		err := handler.AddUserToGroup(ctx, "test-group-3", &req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("user already in group", func(t *testing.T) {
		handler.DBClient.StoreUser(ctx, User{UserId: "test-user-4"})
		handler.DBClient.StoreGroup(ctx, Group{GroupId: "test-group-4"})
		req := UserToGroupRequest{
			UserId: "test-user-4",
		}
		handler.AddUserToGroup(ctx, "test-group-4", &req)

		// add the user again
		err := handler.AddUserToGroup(ctx, "test-group-4", &req)
		assert.Error(t, err)
		assert.IsType(t, &common.BadRequestError{}, err)
	})
}

func TestRemoveUserFromGroup(t *testing.T) {
	ctx := context.Background()
	handler := GroupHandler{
		DBClient: db.NewMockDBClient(),
	}
	t.Run("Remove user from group successfully", func(t *testing.T) {
		handler.DBClient.StoreUser(ctx, User{UserId: "test-user-1"})
		handler.DBClient.StoreGroup(ctx, Group{GroupId: "test-group-1"})
		handler.DBClient.AddUserToGroup(ctx, Group{GroupId: "test-group-1"}, User{UserId: "test-user-1"})

		// create a new request
		req := UserToGroupRequest{
			UserId: "test-user-1",
		}
		err := handler.RemoveUserFromGroup(ctx, "test-group-1", &req)
		assert.NoError(t, err)

	})
}
