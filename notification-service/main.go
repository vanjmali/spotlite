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
	"github.com/vanjmali/spotlite/notifications/repositories"
	"github.com/vanjmali/spotlite/notifications/routers"
	"github.com/vanjmali/spotlite/notifications/services"
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

	// schema initialization
	if err := infrastructure.InitializeSchema(cassHost, ks); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	// initialize cassandra session which will be used to execute queries
	cs, err := infrastructure.Initialize(cassHost, ks)
	if err != nil {
		return fmt.Errorf("failed to initialize database session: %v", err)
	}
	defer cs.Close()

	// initialize notification broker
	b := infrastructure.NewBroker()
	go b.Listen()

	nr := repositories.NewNotificationRepository(cs)
	ns := services.NewNotificationService(nr)
	nh := handlers.NewNotificationHandler(ns, b)
	r := routers.HandleRequests(nh)

	srvAddr := ":" + port

	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	srvErr := make(chan error, 1)

	// stop is a channel which stores a maximum of one os signal
	stop := make(chan os.Signal, 1)

	// when an os.Interupt (ctrl + C) OR Sigterm call occurs, sends a signal to the stop channel
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// starts the http server in a new goroutine so graceful shutdown mechanism doesn't get blocked and can
	// react of signals
	go func() {
		log.Printf("INFO: Listening on %s", srvAddr)
		// If an error occurs, send it to the channel. Do NOT log.Fatal here.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	// stops the line of execution here until the stop channels gets a signal
	select {
	case sig := <-stop:
		log.Printf("DEBUG: Received signal %v. Shutting down gracefully...", sig)
	case err := <-srvErr:
		// if we get here, the server failed to start We log it, but we do NOT exit
		// immediately. We let the function finish so that the 'defer' statements
		// above trigger.
		log.Printf("ERROR: HTTP server failed to start: %v", err)
		return err // return the error to main
	}

	// graceful shutdown starts
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("ERROR: HTTP server Shutdown error: %v", err)
	}

	log.Println("DEBUG: Shutdown complete")

	return nil
}
