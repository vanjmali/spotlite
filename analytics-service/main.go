package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "analytics-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		// TODO: Add custom validators
		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		// TODO: Initialize MongoDB client with ROOT_CERT_PATH, CERT_PATH, KEY_PATH
		// TODO: Initialize NATS client
		// TODO: Initialize repositories
		// TODO: Initialize services
		// TODO: Initialize handlers
		// TODO: Initialize router

		shutdown = func() error {
			return nil
		}

		return nil, shutdown, fmt.Errorf("not yet implemented")
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		panic(err)
	}
}
