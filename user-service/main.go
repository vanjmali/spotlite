package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mailing"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
)

var port = utils.GetEnv("APP_PORT", "3000")

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Initialize telemetry
	tr, err := telemetry.Init(ctx, "user-service")
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	defer func() {
		if err := tr.Shutdown(ctx); err != nil {
			log.Printf("failed to shut down tracer provider: %v", err)
		}
	}()

	// Initialize clients
	dbClient, err := mongo.InitMongoClient()
	if err != nil {
		return fmt.Errorf("cannot start application without DB connection: %w", err)
	}

	mailClient, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = dbClient.Disconnect(context.Background())
		return fmt.Errorf("cannot start application without mailing service: %w", err)
	}

	defer dbClient.Disconnect(context.Background())
	defer mailClient.Close()

	// Configure validators
	val := validator.New()
	if err := requests.RegisterValidation(val, validation.CheckStrongPassword); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	if err := requests.RegisterValidation(val, validation.CheckValidUsername); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	// Initialize repositories, services, handlers, and routers
	userRepo := repositories.NewRepository(mongo.DatabaseName(), "users", dbClient)
	ms := services.InitMailingService(mailClient)
	us := services.NewUserService(*userRepo, *ms)

	rtRepo := repositories.NewRefreshTokenRepository(mongo.DatabaseName(), repositories.RefreshTokensColl, dbClient)
	if err := rtRepo.EnsureRefreshIndexes(context.Background()); err != nil {
		return fmt.Errorf("failed to ensure refresh token indexes: %w", err)
	}

	rts := services.NewRefreshTokenService(*rtRepo)
	userH := handlers.NewUserHandler(*us, *val, *rts)
	rtH := handlers.NewRefreshTokenHandler(*rts, *us, *val)

	router := routers.HandleRequests(userH, rtH)

	// Start HTTP server
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
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
