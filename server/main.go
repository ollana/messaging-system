package main

import (
	"fmt"
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

	http.HandleFunc("/v1/users/register", registerUser)
	http.HandleFunc("/v1/users/:userId/block", blockUser)

	http.HandleFunc("/v1/groups/create", createGroup)
	http.HandleFunc("/v1/groups/:groupId/add", addUserToGroup)
	http.HandleFunc("/v1/groups/:groupId/remove", removeUserFromGroup)

	http.HandleFunc("/v1/messages/group", sendGroupMessage)
	http.HandleFunc("/v1/messages/:userId", getMessages)
	http.HandleFunc("/v1/messages/private", sendPrivateMessage)

	fmt.Println("Starting server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
