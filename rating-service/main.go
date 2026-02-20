package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	commonvalid "github.com/vanjmali/spotlite/common-lib/validations"
	contentvalid "github.com/vanjmali/spotlite/content/validations"
	"github.com/vanjmali/spotlite/ratings/infrastructure/mongo"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "rating-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		requests.RegisterJSONTagNameFunc(v)
		if err := requests.RegisterValidation(v, commonvalid.CheckValidName); err != nil {
			return fmt.Errorf("failed to register name validation: %w", err)
		}
		if err := requests.RegisterValidation(v, contentvalid.CheckValidDateOnly); err != nil {
			return fmt.Errorf("failed to register date-only validation: %w", err)
		}

		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		dbc, err := createClients()
		if err != nil {
			err = fmt.Errorf("failed to create clients: %w", err)
			return h, shutdown, err
		}

		// Cleanup resources on error
		defer func() {
			if err == nil {
				return
			}
			_ = dbc.Disconnect(ctx)
		}()

		shutdown = func() error {

			if err := dbc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}

			return nil
		}

		return h, shutdown, err
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start content service: %v", err)
	}
}

func createClients() (*mongodriver.Client, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	return dbc, nil
}
