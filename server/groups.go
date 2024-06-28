package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"net/http"
)

func createGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var group map[string]string
	err := decoder.Decode(&group)
	if err != nil || group["name"] == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	item := map[string]types.AttributeValue{
		"GroupId": &types.AttributeValueMemberS{Value: group["name"]},
		"Members": &types.AttributeValueMemberSS{Value: []string{}},
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"GroupId": group["name"]}
	json.NewEncoder(w).Encode(resp)
}

func addUserToGroup(w http.ResponseWriter, r *http.Request) {

}

func removeUserFromGroup(w http.ResponseWriter, r *http.Request) {

}
