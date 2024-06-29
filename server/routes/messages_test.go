package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"server/common"
	"server/db"
	"server/messages"
	"testing"
)

type messageHandlerMock struct {
	error error
}

func (mh *messageHandlerMock) SendPrivateMessage(ctx context.Context, req messages.SendMessageRequest) error {
	if mh.error != nil {
		return mh.error
	}
	return nil
}
func (mh *messageHandlerMock) SendGroupMessage(ctx context.Context, req messages.SendMessageRequest) error {
	if mh.error != nil {
		return mh.error
	}
	return nil
}

func (mh *messageHandlerMock) GetMessages(ctx context.Context, recipientId string, timestamp int64) (*messages.UserMessagesResp, error) {
	if mh.error != nil {
		return nil, mh.error
	}
	return &messages.UserMessagesResp{Messages: []db.Message{db.Message{Message: "hello"}}}, nil
}

func TestSendMessageHandler(t *testing.T) {
	r := Router{Messages: MessagesRoutes{Handler: &messageHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Happy path private msg", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/messages/send?type=private", bytes.NewReader([]byte(`{"SenderId": "sender", "RecipientId": "recipient", "Message": "hello"}`)))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Happy path group msg", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/messages/send?type=group", bytes.NewReader([]byte(`{"SenderId": "sender", "RecipientId": "recipient", "Message": "hello"}`)))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid input", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/messages/send?type=private", bytes.NewReader([]byte(`{}`)))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/messages/send?type=invalid", bytes.NewReader([]byte(`{"SenderId": "sender", "RecipientId": "recipient", "Message": "hello"}`)))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		r := Router{Messages: MessagesRoutes{Handler: &messageHandlerMock{error: &common.ForbiddenError{Message: "error"}}}}
		router, err := r.NewRouter()
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/messages/send?type=private", bytes.NewReader([]byte(`{"SenderId": "sender", "RecipientId": "recipient", "Message": "hello"}`)))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

}

func TestGetMessagesHandler(t *testing.T) {
	r := Router{Messages: MessagesRoutes{Handler: &messageHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Happy path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/v1/messages/recipient", nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		decoder := json.NewDecoder(w.Body)
		var resp messages.UserMessagesResp
		decoder.Decode(&resp)
		assert.Equal(t, "hello", resp.Messages[0].Message)
	})

	t.Run("Error", func(t *testing.T) {
		r := Router{Messages: MessagesRoutes{Handler: &messageHandlerMock{error: &common.NotFoundError{Message: "error"}}}}
		router, err := r.NewRouter()
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/v1/messages/recipient", nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
