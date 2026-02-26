package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
)

var (
	rootCACertFilePath = utils.MustGetEnv("ROOT_CERT_PATH")
	certFilePath       = utils.MustGetEnv("CERT_PATH")
	keyFilePath        = utils.MustGetEnv("KEY_PATH")
	config             = server.ServerRunConfiguration{
		TelemetryName: "analytics-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		ConfigureValidation: func(v *validator.Validate) error {
			// TODO: Add custom validators
			return nil
		},
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			// TODO: Initialize MongoDB client
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
)

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		panic(err)
	}
}
