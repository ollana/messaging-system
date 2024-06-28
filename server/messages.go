package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"net/http"
	"time"
)

func sendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msg map[string]string
	err := decoder.Decode(&msg)
	if err != nil || msg["senderId"] == "" || msg["recipientId"] == "" || msg["message"] == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	item := map[string]types.AttributeValue{
		"RecipientId": &types.AttributeValueMemberS{Value: msg["recipientId"]},
		"Timestamp":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		"SenderId":    &types.AttributeValueMemberS{Value: msg["senderId"]},
		"Message":     &types.AttributeValueMemberS{Value: msg["message"]},
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendGroupMessage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msg map[string]string
	err := decoder.Decode(&msg)
	if err != nil || msg["groupId"] == "" || msg["senderId"] == "" || msg["message"] == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	item := map[string]types.AttributeValue{
		"RecipientId": &types.AttributeValueMemberS{Value: msg["groupId"]},
		"Timestamp":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		"SenderId":    &types.AttributeValueMemberS{Value: msg["senderId"]},
		"Message":     &types.AttributeValueMemberS{Value: msg["message"]},
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	recipientId := r.URL.Query().Get("recipientId")
	if recipientId == "" {
		http.Error(w, "recipientId is required", http.StatusBadRequest)
		return
	}

	// Query DynamoDB to get messages for the recipient
	result, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("RecipientId = :recipientId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":recipientId": &types.AttributeValueMemberS{Value: recipientId},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect messages from the query result
	messages := []map[string]string{}
	for _, item := range result.Items {
		msg := map[string]string{
			"senderId":  item["SenderId"].(*types.AttributeValueMemberS).Value,
			"message":   item["Message"].(*types.AttributeValueMemberS).Value,
			"timestamp": item["Timestamp"].(*types.AttributeValueMemberN).Value,
		}
		messages = append(messages, msg)

	}

	// Return the messages as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
