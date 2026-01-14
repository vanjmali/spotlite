package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/notifications/handlers"
	"github.com/vanjmali/spotlite/notifications/infrastructure"
	"github.com/vanjmali/spotlite/notifications/routers"
)

var (
	port     = utils.GetEnv("APP_PORT", "3000")
	cassHost = utils.GetEnv("CASSANDRA_HOST", "127.0.0.1")
	ks       = "notification_service"
)

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

	if err := infrastructure.CreateKeyspace(cassHost, ks); err != nil {
		log.Fatalf("Failed to create keyspace: %v", err)
	}

	cs, err := infrastructure.Initialize(cassHost, ks)
	if err != nil {
		return fmt.Errorf("failed to initialize database session: %v", err)
	}
	defer cs.Close()

	var version string
	if err := cs.Query("SELECT release_version FROM system.local").Scan(&version); err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}

	fmt.Printf("✅ Connection Successful! Cassandra Version: %s\n", version)

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

	// stop is a channel which stores a maximum of one os signal
	stop := make(chan os.Signal, 1)

	// when an os.Interupt (ctrl + C) OR Sigterm call occurs, sends a signal to the stop channel
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// starts the http server in a new goroutine so graceful shutdown mechanism doesn't get blocked and can
	// react of signals
	go func() {
		log.Printf("INFO: Listening on %s", srvAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ERROR: failed to start server: %s", err)
		}
	}()

	// stops the line of execution here until the stop channels gets a signal
	<-stop
	log.Println("DEBUG: Shutting down gracefully...")

	// graceful shutdown starts
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("ERROR: HTTP server Shutdown error: %v", err)
	}

	log.Println("DEBUG: Shutdown complete")

	return nil
}
