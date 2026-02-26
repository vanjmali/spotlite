package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/analytics-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "analytics-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		// TODO: Add custom validators
		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		mc, err := createClients()
		if err != nil {
			return h, shutdown, fmt.Errorf("failed to create clients: %w", err)
		}

		// Cleanup resources on error
		defer func() {
			if err == nil {
				return
			}
			_ = mc.Disconnect(ctx)
		}()

		// TODO: Initialize NATS client
		// TODO: Initialize repositories
		// TODO: Initialize services
		// TODO: Initialize handlers
		// TODO: Initialize router

		shutdown = func() error {
			if err := mc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}
			return nil
		}

		return nil, shutdown, fmt.Errorf("not yet implemented")
	},
}

func createClients() (*mongodriver.Client, error) {
	mc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}
	return mc, nil
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		panic(err)
	}
}
