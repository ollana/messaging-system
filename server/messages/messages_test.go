package messages

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"server/common"
	. "server/common"
	"server/db"
	"testing"
	"time"
)

func TestSendPrivateMessage(t *testing.T) {
	ctx := context.Background()

	handler := Handler{DBClient: db.NewMockDBClient()}

	t.Run("Send private message successfully", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user1)
		handler.DBClient.StoreUser(ctx, user2)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user1.UserId,
			RecipientId: user2.UserId,
			Message:     "Hello",
		}

		err := handler.SendPrivateMessage(ctx, req)
		assert.NoError(t, err)

		assert.NotEmpty(t, handler.DBClient.(*db.MockDBClient).Messages[req.RecipientId])
		// assert message
		msg := handler.DBClient.(*db.MockDBClient).Messages[req.RecipientId][0]
		assert.Equal(t, req.SenderId, msg.SenderId)
		assert.Equal(t, req.Message, msg.Message)

	})

	t.Run("User not found", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user1)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user1.UserId,
			RecipientId: user2.UserId,
			Message:     "Hello",
		}

		err := handler.SendPrivateMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("db error", func(t *testing.T) {
		handler := Handler{DBClient: db.NewMockDBClient()}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := SendMessageRequest{
			SenderId:    "test-user-1",
			RecipientId: "test-user-2",
			Message:     "Hello",
		}
		err := handler.SendPrivateMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)
	})

}

func TestSendGroupMessage(t *testing.T) {
	ctx := context.Background()
	handler := Handler{DBClient: db.NewMockDBClient()}

	t.Run("Send group message successfully", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user)
		handler.DBClient.StoreGroup(ctx, group)
		handler.DBClient.AddUserToGroup(ctx, group, user)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user.UserId,
			RecipientId: group.GroupId,
			Message:     "Hello",
		}

		err := handler.SendGroupMessage(ctx, req)
		assert.NoError(t, err)

		assert.NotEmpty(t, handler.DBClient.(*db.MockDBClient).Messages[req.RecipientId])
		// assert message
		msg := handler.DBClient.(*db.MockDBClient).Messages[req.RecipientId][0]
		assert.Equal(t, req.SenderId, msg.SenderId)
		assert.Equal(t, req.Message, msg.Message)

	})

	t.Run("Sender not a member of the group", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user)
		handler.DBClient.StoreGroup(ctx, group)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user.UserId,
			RecipientId: group.GroupId,
			Message:     "Hello",
		}

		err := handler.SendGroupMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.ForbiddenError{}, err)
	})

	t.Run("Group not found", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user.UserId,
			RecipientId: group.GroupId,
			Message:     "Hello",
		}

		err := handler.SendGroupMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("Sender not found", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		handler.DBClient.StoreGroup(ctx, group)

		// create a new request
		req := SendMessageRequest{
			SenderId:    user.UserId,
			RecipientId: group.GroupId,
			Message:     "Hello",
		}

		err := handler.SendGroupMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.NotFoundError{}, err)
	})

	t.Run("db error", func(t *testing.T) {
		handler := Handler{DBClient: db.NewMockDBClient()}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		req := SendMessageRequest{
			SenderId:    "test-user-1",
			RecipientId: "test-group-1",
			Message:     "Hello",
		}
		err := handler.SendGroupMessage(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)
	})

}

func TestGetMessages(t *testing.T) {
	ctx := context.Background()
	handler := Handler{DBClient: db.NewMockDBClient()}

	t.Run("Get empty messages successfully", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		handler.DBClient.StoreUser(ctx, user)
		handler.DBClient.StoreGroup(ctx, group)
		handler.DBClient.AddUserToGroup(ctx, group, user)

		msgs, err := handler.GetMessages(ctx, user.UserId, 0)
		assert.NoError(t, err)
		assert.Empty(t, msgs.Messages)

	})

	t.Run("Get direct messages successfully", func(t *testing.T) {
		user1 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		user2 := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}

		msg := Message{
			RecipientId: user1.UserId,
			Timestamp:   time.Now().Add(-time.Hour).Format(time.RFC3339),
			SenderId:    user2.UserId,
			Message:     "hello",
		}
		handler.DBClient.StoreUser(ctx, user1)
		handler.DBClient.StoreUser(ctx, user2)
		handler.DBClient.StoreMessage(ctx, msg)

		t.Run("no timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user1.UserId, 0)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(msgs.Messages))
			assert.Contains(t, msgs.Messages, msg)
		})

		t.Run("with timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user1.UserId, time.Now().Add(-2*time.Hour).Unix())
			assert.NoError(t, err)
			assert.Equal(t, 1, len(msgs.Messages))
			assert.Contains(t, msgs.Messages, msg)
		})

		t.Run("with greater timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user1.UserId, time.Now().Unix())
			assert.NoError(t, err)
			assert.Equal(t, 0, len(msgs.Messages))
		})

	})

	t.Run("Get group messages successfully", func(t *testing.T) {
		user := User{UserId: fmt.Sprintf("test-user-%s", uuid.New().String())}
		group := Group{GroupId: fmt.Sprintf("test-group-%s", uuid.New().String())}

		msg := Message{
			RecipientId: group.GroupId,
			Timestamp:   time.Now().Add(-time.Hour).Format(time.RFC3339),
			SenderId:    user.UserId,
			Message:     "hello",
		}
		handler.DBClient.StoreUser(ctx, user)
		handler.DBClient.StoreGroup(ctx, group)
		handler.DBClient.AddUserToGroup(ctx, group, user)
		handler.DBClient.StoreMessage(ctx, msg)

		t.Run("no timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user.UserId, 0)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(msgs.Messages))
			assert.Contains(t, msgs.Messages, msg)
		})

		t.Run("with timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user.UserId, time.Now().Add(-2*time.Hour).Unix())
			assert.NoError(t, err)
			assert.Equal(t, 1, len(msgs.Messages))
			assert.Contains(t, msgs.Messages, msg)
		})

		t.Run("with greater timestamp", func(t *testing.T) {
			msgs, err := handler.GetMessages(ctx, user.UserId, time.Now().Unix())
			assert.NoError(t, err)
			assert.Equal(t, 0, len(msgs.Messages))
		})
	})

	t.Run("db error", func(t *testing.T) {
		handler := Handler{DBClient: db.NewMockDBClient()}

		handler.DBClient.(*db.MockDBClient).Error = fmt.Errorf("some error")
		_, err := handler.GetMessages(ctx, "user-1", 0)
		assert.Error(t, err)
		assert.IsType(t, &common.InternalServerError{}, err)
	})

}
