package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"server/common"
	"server/groups"
	"testing"
)

type groupHandlerMock struct {
	error error
}

func (gh *groupHandlerMock) CreateGroup(ctx context.Context, req *groups.CreateGroupRequest) (*groups.CreateGroupResponse, error) {
	if gh.error != nil {
		return nil, gh.error
	}
	return &groups.CreateGroupResponse{GroupId: req.GroupName, GroupName: req.GroupName}, nil
}
func (gh *groupHandlerMock) AddUserToGroup(ctx context.Context, groupId string, req *groups.UserToGroupRequest) error {
	if gh.error != nil {
		return gh.error
	}
	return nil
}
func (gh *groupHandlerMock) RemoveUserFromGroup(ctx context.Context, groupId string, req *groups.UserToGroupRequest) error {
	if gh.error != nil {
		return gh.error
	}
	return nil
}

func TestCreateGroupHandler(t *testing.T) {
	r := Router{Groups: GroupRoutes{Handler: &groupHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Create successfully", func(t *testing.T) {

		reqBody := groups.CreateGroupRequest{
			GroupName: "test-group",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		decoder := json.NewDecoder(w.Body)
		var groupRes groups.CreateGroupResponse
		_ = decoder.Decode(&groupRes)

		assert.Equal(t, reqBody.GroupName, groupRes.GroupName)

	})

	t.Run("Invalid input", func(t *testing.T) {
		reqBody := groups.CreateGroupRequest{
			GroupName: "",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		r := Router{Groups: GroupRoutes{Handler: &groupHandlerMock{error: assert.AnError}}}
		router, err := r.NewRouter()
		assert.Nil(t, err)

		reqBody := groups.CreateGroupRequest{
			GroupName: "test-group",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/create", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

}

func TestUserToGroupHandler(t *testing.T) {
	r := Router{Groups: GroupRoutes{Handler: &groupHandlerMock{}}}
	router, err := r.NewRouter()
	assert.Nil(t, err)

	t.Run("Add user successfully", func(t *testing.T) {
		reqBody := groups.UserToGroupRequest{
			UserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/test-group?op=add", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Remove user successfully", func(t *testing.T) {
		reqBody := groups.UserToGroupRequest{
			UserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/test-group?op=remove", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid input", func(t *testing.T) {
		reqBody := groups.UserToGroupRequest{
			UserId: "",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/test-group?op=add", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not found error", func(t *testing.T) {
		r := Router{Groups: GroupRoutes{Handler: &groupHandlerMock{error: &common.NotFoundError{Message: "some error"}}}}
		router, err := r.NewRouter()
		assert.Nil(t, err)

		reqBody := groups.UserToGroupRequest{
			UserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/?op=add", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid op", func(t *testing.T) {
		reqBody := groups.UserToGroupRequest{
			UserId: "test-user",
		}
		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/v1/groups/test-group?op=invalid", bytes.NewReader(body))
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
