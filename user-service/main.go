package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
)

func main() {
	client, err := mongo.InitMongoClient()
	if err != nil {
		log.Fatalf("FATAL: Cannot start application without DB connection: %v", err)
	}
	defer client.Disconnect(context.Background())

	val := validator.New()
	err = val.RegisterValidation("strongpassword", validation.CheckStrongPassword)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	err = val.RegisterValidation("validusername", validation.CheckValidUsername)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	repo := repositories.NewRepository(mongo.DatabaseName(), "users", client)
	service := services.NewUserService(*repo)
	handler := handlers.NewUserHandler(*service, *val)

	router := routers.HandleRequests(handler)

	log.Println("Server starting on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
