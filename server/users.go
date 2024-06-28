package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func registerUser(w http.ResponseWriter, r *http.Request) {
	userId := fmt.Sprintf("user-%d", time.Now().UnixNano())
	resp := map[string]string{"UserId": userId}
	json.NewEncoder(w).Encode(resp)
}

func blockUser(w http.ResponseWriter, r *http.Request) {

}

func isBlockedUser(w http.ResponseWriter, r *http.Request) {

}
