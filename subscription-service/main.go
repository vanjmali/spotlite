package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/subscriptions/routers"
)

var port = utils.GetEnv("APP_PORT", "3000")

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: Couldn't start subscription service: %v", err)
	}
}

func run() error {

	ctx := context.Background()

	tr, err := telemetry.Init(ctx, "subscription-service")
	if err != nil {
		return fmt.Errorf("failed to initialize subscription service tracing: %w", err)
	}

	defer func() {
		if err := tr.Shutdown(ctx); err != nil {
			log.Printf("failed to shut down subscription service tracer provider: %v", err)
		}
	}()

	r := routers.HandleRequests()

	srvAddr := ":" + port

	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("INFO: Listening on %s", srvAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ERROR: failed to start server: %s", err)
	}

	return nil
}
