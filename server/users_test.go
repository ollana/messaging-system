package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"server/common"
	"server/db"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	dbClient = db.NewMockDBClient()

	t.Run("Register user successfully", func(t *testing.T) {
		// create a new request
		req := RegisterUserRequest{
			UserName: "test-user",
		}
		resp, err := registerUser(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.UserId)
		assert.Equal(t, req.UserName, resp.UserName)

	})

	t.Run("db error", func(t *testing.T) {
		dbClient = db.NewMockDBClient()
		dbClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := RegisterUserRequest{
			UserName: "test-user",
		}
		resp, err := registerUser(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)
		assert.Nil(t, resp)

	})

}

func TestBlockUser(t *testing.T) {
	ctx := context.Background()
	dbClient = db.NewMockDBClient()

	t.Run("Block user successfully", func(t *testing.T) {
		user1 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		dbClient.StoreUser(context.Background(), user1)
		dbClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := blockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

	})

	t.Run("non existing user", func(t *testing.T) {
		// create a new request
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := blockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("already blocked", func(t *testing.T) {
		user1 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		dbClient.StoreUser(context.Background(), user1)
		dbClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := blockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

		err = blockUser(ctx, user1.UserId, req)
		assert.Error(t, err)
		assert.IsType(t, &common.BadRequestError{}, err)

	})

	t.Run("db error", func(t *testing.T) {
		dbClient = db.NewMockDBClient()
		dbClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := blockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)

	})

}

func TestUnblockUser(t *testing.T) {
	ctx := context.Background()
	dbClient = db.NewMockDBClient()

	t.Run("Unblock user successfully", func(t *testing.T) {
		user1 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		dbClient.StoreUser(context.Background(), user1)
		dbClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := blockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

		err = unblockUser(ctx, user1.UserId, req)
		assert.NoError(t, err)

	})

	t.Run("non existing user", func(t *testing.T) {
		// create a new request
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := unblockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("not blocked", func(t *testing.T) {
		user1 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := db.User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		dbClient.StoreUser(context.Background(), user1)
		dbClient.StoreUser(context.Background(), user2)

		// create a new request
		req := BlockUserRequest{
			BlockedUserId: user2.UserId,
		}

		err := unblockUser(ctx, user1.UserId, req)
		assert.Error(t, err)
		assert.IsType(t, &common.BadRequestError{}, err)
	})

	t.Run("db error", func(t *testing.T) {
		dbClient = db.NewMockDBClient()
		dbClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := BlockUserRequest{
			BlockedUserId: "test-user-2",
		}

		err := unblockUser(ctx, "test-user-1", req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)

	})
}
