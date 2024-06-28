package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/uber/jaeger-client-go/log/zap"
	"log"
	"net/http"
)

var dbClient dynamoDBClientInterface
var logger zap.Logger

func main() {
	var err error
	dbClient, err = NewDynamoDBClient()
	if err != nil {
		log.Fatalf("Error creating DynamoDB client, %v", err)
	}

	r := chi.NewRouter()
	r.Post("/v1/users/register", registerUser)
	r.Post("/v1/users/{userId}/block", blockUser)

	r.Post("/v1/groups/create", createGroup)
	r.Post("/v1/groups/{groupId}/add", addUserToGroup)
	r.Post("/v1/groups/{groupId}/remove", removeUserFromGroup)

	r.Post("/v1/messages/group", sendGroupMessage)
	r.Get("/v1/messages/{userId}", getMessages)
	r.Post("/v1/messages/private", sendPrivateMessage)

	fmt.Println("Starting server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
