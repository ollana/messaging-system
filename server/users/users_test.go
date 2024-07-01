package users

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	. "server/common"
	"server/db"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	handler := UsersHandler{DBClient: db.NewMockDBClient()}

	t.Run("Register user successfully", func(t *testing.T) {
		// create a new request
		req := RegisterUserRequest{
			UserName: "test-user",
		}
		resp, err := handler.RegisterUser(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.UserId)
		assert.Equal(t, req.UserName, resp.UserName)

	})

	t.Run("db error", func(t *testing.T) {
		handler := UsersHandler{DBClient: db.NewMockDBClient()}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := RegisterUserRequest{
			UserName: "test-user",
		}
		resp, err := handler.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &InternalServerError{}, err)
		assert.Nil(t, resp)

	})

}

func TestBlockUser(t *testing.T) {
	ctx := context.Background()
	handler := UsersHandler{DBClient: db.NewMockDBClient()}

	t.Run("Block user successfully", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(context.Background(), user1)
		handler.DBClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := handler.BlockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

	})

	t.Run("non existing user", func(t *testing.T) {
		// create a new request
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := handler.BlockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &NotFoundError{}, err)
	})

	t.Run("already blocked", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(context.Background(), user1)
		handler.DBClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := handler.BlockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

		err = handler.BlockUser(ctx, user1.UserId, req)
		assert.Error(t, err)
		assert.IsType(t, &BadRequestError{}, err)

	})

	t.Run("db error", func(t *testing.T) {
		handler := UsersHandler{DBClient: db.NewMockDBClient()}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := handler.BlockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &InternalServerError{}, err)

	})

}

func TestUnblockUser(t *testing.T) {
	ctx := context.Background()
	handler := UsersHandler{DBClient: db.NewMockDBClient()}

	t.Run("Unblock user successfully", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(context.Background(), user1)
		handler.DBClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := handler.BlockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

		err = handler.UnblockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

	})

	t.Run("non existing user", func(t *testing.T) {
		// create a new request
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := handler.UnblockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &NotFoundError{}, err)
	})

	t.Run("not blocked", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(context.Background(), user1)
		handler.DBClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := handler.UnblockUser(ctx, user1.UserId, req)
		assert.Error(t, err)
		assert.IsType(t, &BadRequestError{}, err)
	})

	t.Run("db error", func(t *testing.T) {
		handler.DBClient = db.NewMockDBClient()
		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := handler.UnblockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &InternalServerError{}, err)

	})
}
