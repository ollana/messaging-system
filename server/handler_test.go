package main

//import (
//	"bytes"
//	"github.com/go-chi/chi/v5"
//	"github.com/stretchr/testify/assert"
//	"net/http/httptest"
//)
//
//func TestRegisterUserHandler(t *testing.T) {
//
//	t.Run("Register user successfully", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		// create a new request
//		req := RegisterUserRequest{
//			UserName: "test-user",
//		}
//		body, _ := json.Marshal(req)
//		r := httptest.NewRequest("POST", "/v1/users/register", bytes.NewReader(body))
//		w := httptest.NewRecorder()
//		// aseed the request to the handler
//		registerUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 200, w.Code)
//		// check the response body
//		var resp RegisterUserResponse
//		json.NewDecoder(w.Body).Decode(&resp)
//		assert.NotEmpty(t, resp.UserId)
//		assert.Equal(t, req.UserName, resp.UserName)
//
//	})
//
//	t.Run("Invalid input", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		// create a new request
//		req := RegisterUserRequest{
//			UserName: "",
//		}
//		body, _ := json.Marshal(req)
//		r := httptest.NewRequest("POST", "/v1/users/register", bytes.NewReader(body))
//		w := httptest.NewRecorder()
//		// aseed the request to the handler
//		registerUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 400, w.Code)
//	})
//
//	t.Run("db error", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		dbClient.(*mockDBClient).error = fmt.Errorf("some error")
//		req := RegisterUserRequest{
//			UserName: "testuser",
//		}
//		body, _ := json.Marshal(req)
//		r := httptest.NewRequest("POST", "/v1/users/register", bytes.NewReader(body))
//		w := httptest.NewRecorder()
//		// aseed the request to the handler
//		registerUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 500, w.Code)
//	})
//
//}
//
//func TestBlockUserHandler(t *testing.T) {
//	t.Run("Block user successfully", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		dbClient.StoreUser(context.Background(), dbUser{UserId: "test-user-1"})
//		dbClient.StoreUser(context.Background(), dbUser{UserId: "test-user-2"})
//
//		// create a new request
//		req := BlockUserRequest{
//			BlockedUserId: "test-user-2",
//		}
//		body, _ := json.Marshal(req)
//		w := httptest.NewRecorder()
//		r := httptest.NewRequest("POST", "/v1/users/{userId}/block", bytes.NewReader(body))
//
//		rctx := chi.NewRouteContext()
//		rctx.URLParams.Add("userId", "test-user-1")
//		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
//
//		// aseed the request to the handler
//		blockUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 200, w.Code, w.Body.String())
//	})
//
//	t.Run("Invalid input", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		// create a new request
//		req := BlockUserRequest{
//			BlockedUserId: "",
//		}
//		body, _ := json.Marshal(req)
//		w := httptest.NewRecorder()
//		r := httptest.NewRequest("POST", "/v1/users/{userId}/block", bytes.NewReader(body))
//		rctx := chi.NewRouteContext()
//		rctx.URLParams.Add("userId", "")
//
//		blockUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 400, w.Code, w.Body.String())
//	})
//	t.Run("non existing user", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		// create a new request
//		req := BlockUserRequest{
//			BlockedUserId: "test-user-2",
//		}
//		body, _ := json.Marshal(req)
//		w := httptest.NewRecorder()
//		r := httptest.NewRequest("POST", "/v1/users/{userId}/block", bytes.NewReader(body))
//
//		rctx := chi.NewRouteContext()
//		rctx.URLParams.Add("userId", "test-user-1")
//		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
//
//		// aseed the request to the handler
//		blockUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 404, w.Code, w.Body.String())
//	})
//	t.Run("db error", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		dbClient.(*mockDBClient).error = fmt.Errorf("some error")
//		req := BlockUserRequest{
//			BlockedUserId: "test-user-2",
//		}
//		body, _ := json.Marshal(req)
//		w := httptest.NewRecorder()
//		r := httptest.NewRequest("POST", "/v1/users/{userId}/block", bytes.NewReader(body))
//		rctx := chi.NewRouteContext()
//		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
//
//		rctx.URLParams.Add("userId", "test-user-1")
//
//		blockUserHandler(w, r)
//		// check the response status code
//		assert.Equal(t, 500, w.Code, w.Body.String())
//	})
//
//}
