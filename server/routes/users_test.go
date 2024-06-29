package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"server/users"
	"testing"
)

type userHandlerMock struct {
	error error
}

func (uh *userHandlerMock) RegisterUser(ctx context.Context, req users.RegisterUserRequest) (*users.RegisterUserResponse, error) {
	if uh.error != nil {
		return nil, uh.error
	}
	return &users.RegisterUserResponse{UserId: req.UserName, UserName: req.UserName}, nil
}

func (uh *userHandlerMock) BlockUser(ctx context.Context, userId string, req users.BlockUserRequest) error {
	if uh.error != nil {
		return uh.error
	}
	return nil
}

func (uh *userHandlerMock) UnblockUser(ctx context.Context, userId string, req users.BlockUserRequest) error {
	if uh.error != nil {
		return uh.error
	}
	return nil
}

func TestRegisterUserHandler(t *testing.T) {

	r := Router{Users: UsersRoutes{Handler: &userHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Happy path", func(t *testing.T) {
		reqBody := users.RegisterUserRequest{
			UserName: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid input", func(t *testing.T) {
		reqBody := users.RegisterUserRequest{
			UserName: "",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		r := Router{Users: UsersRoutes{Handler: &userHandlerMock{error: assert.AnError}}}
		router, err := r.NewRouter()
		assert.Nil(t, err)

		reqBody := users.RegisterUserRequest{
			UserName: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBlockUserHandler(t *testing.T) {
	r := Router{Users: UsersRoutes{Handler: &userHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Block happy path", func(t *testing.T) {
		reqBody := users.BlockUserRequest{
			BlockedUserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/test-user?op=block", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unblock happy path", func(t *testing.T) {
		reqBody := users.BlockUserRequest{
			BlockedUserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/test-user?op=unblock", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid op", func(t *testing.T) {
		reqBody := users.BlockUserRequest{
			BlockedUserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/test-user?op=invalid", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

	})

	t.Run("Invalid input", func(t *testing.T) {
		reqBody := users.BlockUserRequest{
			BlockedUserId: "",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/users/test-user?op=block", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

}
