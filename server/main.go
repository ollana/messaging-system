package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"net/http"
)

var svc *dynamodb.Client
var tableName = "messagesTable"

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc = dynamodb.NewFromConfig(cfg)

	http.HandleFunc("/v1/users/register", registerUser)
	http.HandleFunc("/v1/users/:userId/block", blockUser)
	http.HandleFunc("/v1/users/:userId/blocked/:blockedUserId", isBlockedUser)

	http.HandleFunc("/v1/groups/create", createGroup)
	http.HandleFunc("/v1/groups/:groupId/add", addUserToGroup)
	http.HandleFunc("/v1/groups/:groupId/remove", removeUserFromGroup)

	http.HandleFunc("/v1/messages/group", sendGroupMessage)
	http.HandleFunc("/v1/messages/:userId", getMessages)
	http.HandleFunc("/v1/messages/private", sendPrivateMessage)

	fmt.Println("Starting server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
