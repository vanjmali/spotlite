package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mailing"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/utils"
	"github.com/vanjmali/spotlite/user-service/validation"
)

var port = utils.GetEnv("APP_PORT", "8000")

func main() {
	dbClient, err := mongo.InitMongoClient()
	if err != nil {
		log.Fatalf("FATAL: Cannot start application without DB connection: %v", err)
	}
	defer dbClient.Disconnect(context.Background())

	mailClient, err := mailing.InitClientFromEnv()
	if err != nil {
		log.Fatalf("FATAL: Cannot start application without mailing service: %v", err)
	}
	defer mailClient.Close()

	val := validator.New()
	err = val.RegisterValidation("strongpassword", validation.CheckStrongPassword)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	err = val.RegisterValidation("validusername", validation.CheckValidUsername)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	repo := repositories.NewRepository(mongo.DatabaseName(), "users", dbClient)
	ms := services.InitMailingService(mailClient)
	us := services.NewUserService(*repo, *ms)
	h := handlers.NewUserHandler(*us, *val)

	router := routers.HandleRequests(h)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
