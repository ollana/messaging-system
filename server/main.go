package main

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"log"
	"net/http"
)

var dbClient dynamoDBClientInterface

func main() {
	var err error
	dbClient, err = NewDynamoDBClient()
	if err != nil {
		log.Fatalf("Error creating DynamoDB client, %v", err)
	}

	r := chi.NewRouter()
	r.Post("/v1/users/register", registerUserHandler)
	r.Post("/v1/users/{userId}/{op}", blockUserHandler)

	r.Post("/v1/groups/create", createGroupHandler)
	r.Post("/v1/groups/{groupId}/{op}", addUserToGroupHandler)

	r.Post("/v1/messages/{type}", sendMessageHandler)
	r.Get("/v1/messages/{userId}", getMessagesHandler)

	slog.Info("Starting server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
