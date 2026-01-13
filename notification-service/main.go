package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/notifications/handlers"
	"github.com/vanjmali/spotlite/notifications/routers"
)

var port = utils.GetEnv("APP_PORT", "3000")

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: Couldn't start notification service: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Initialize notification service telemetry
	tr, err := telemetry.Init(ctx, "notification-service")
	if err != nil {
		return fmt.Errorf("failed to initialize notification service tracing: %w", err)
	}

	// Schedule telemetry shutdown when exiting function
	defer func() {
		if err := tr.Shutdown(ctx); err != nil {
			log.Printf("failed to shut down notification service tracer provider: %v", err)
		}
	}()

	h := handlers.NewNotificationHandler()
	r := routers.HandleRequests(h)

	srvAddr := ":" + port

	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Print("up and running")
	srv.ListenAndServe()

	return nil
}
