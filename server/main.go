package main

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"log"
	"net/http"
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

	r := chi.NewRouter()
	r.Post("/v1/users/register", userRoute.RegisterUserHandler)
	r.Post("/v1/users/{userId}/{op}", userRoute.BlockUserHandler)

	r.Post("/v1/groups/create", groupRoute.CreateGroupHandler)
	r.Post("/v1/groups/{groupId}/{op}", groupRoute.UserToGroupHandler)

	r.Post("/v1/messages/{type}", messageRoute.SendMessageHandler)
	r.Get("/v1/messages/{userId}", messageRoute.GetMessagesHandler)

	slog.Info("Starting server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
