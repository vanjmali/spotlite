package main

import (
	"context"
	"log"
	"net/http"
	"time"

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

var port = utils.GetEnv("APP_PORT", "3000")

func main() {
	dbClient, err := mongo.InitMongoClient()
	if err != nil {
		log.Fatalf("FATAL: Cannot start application without DB connection: %v", err)
	}

	mailClient, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = dbClient.Disconnect(context.Background())
		log.Fatalf("FATAL: Cannot start application without mailing service: %v", err)
	}
	val := validator.New()
	err = val.RegisterValidation("strongpassword", validation.CheckStrongPassword)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	err = val.RegisterValidation("validusername", validation.CheckValidUsername)
	if err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	defer dbClient.Disconnect(context.Background())
	defer mailClient.Close()

	repo := repositories.NewRepository(mongo.DatabaseName(), "users", dbClient)
	ms := services.InitMailingService(mailClient)
	us := services.NewUserService(*userRepo, *ms)

	rtRepo := repositories.NewRefreshTokenRepository(mongo.DatabaseName(), repositories.RefreshTokensColl, dbClient)
	if err := rtRepo.EnsureRefreshIndexes(context.Background()); err != nil {
		log.Fatalf("Failed to ensure refresh token indexes: %v", err)
	}
	rts := services.NewRefreshTokenService(*rtRepo)
	userH := handlers.NewUserHandler(*us, *val, *rts)
	rtH := handlers.NewRefreshTokenHandler(*rts, *us)

	router := routers.HandleRequests(userH, rtH)

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server failed: %v", err)
	}
}
