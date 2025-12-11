package main

import (
	"context"
	"log"
	"net/http"

	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
)

func main() {
	client, err := mongo.InitMongoClient()
	if err != nil {
		log.Fatalf("FATAL: Cannot start application without DB connection: %v", err)
	}
	defer client.Disconnect(context.Background())

	repo := repositories.NewRepository("user_service_db", "users", client)
	service := services.NewUserService(*repo)
	handler := handlers.NewUserHandler(*service)

	router := routers.HandleRequests(handler)

	log.Println("Server starting on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
