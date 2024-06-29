package main

import (
	"log"
	"server/db"
	"server/groups"
	"server/messages"
	"server/routes"
	"server/users"
)

var dbClient db.DynamoDBClientInterface

func main() {
	var err error
	dbClient, err = db.NewDynamoDBClient()
	if err != nil {
		log.Fatalf("Error creating DynamoDB client, %v", err)
	}

	groupRoute := routes.GroupRoutes{
		Handler: &groups.GroupHandler{DBClient: dbClient},
	}
	userRoute := routes.UsersRoutes{
		Handler: &users.UsersHandler{DBClient: dbClient},
	}
	messageRoute := routes.MessagesRoutes{
		Handler: &messages.Handler{DBClient: dbClient},
	}

	r := routes.Router{
		Users:    userRoute,
		Groups:   groupRoute,
		Messages: messageRoute,
	}
	router, err := r.NewRouter()

	router.Run(":8080")
}
