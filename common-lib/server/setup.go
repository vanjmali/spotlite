package server

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

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"golang.org/x/sync/errgroup"
)

type ServerRunConfiguration struct {
	TelemetryName       string
	Port                string
	ConfigureValidation func(v *validator.Validate) error
	CreateHandler       func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error)
	// GracefulShutdownTimeout is the maximum amount of time to wait for the server to shutdown gracefully. If zero, a default of 10 seconds is used.
	GracefulShutdownTimeout time.Duration
	// Server configuration overrides. Defaults will be used for any zero values.
	Server struct {
		// ReadTimeout is the maximum duration for reading the entire request, including the body. Default is 15 seconds.
		ReadTimeout time.Duration
		// WriteTimeout is the maximum duration before timing out writes of the response. Default is 15 seconds.
		WriteTimeout time.Duration
		// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled. Default is 60 seconds.
		IdleTimeout time.Duration
	}
}

// Run starts the HTTP server based on the provided configuration.
// It handles graceful shutdown on receiving termination signals.
func Run(ctx context.Context, config ServerRunConfiguration) error {
	if err := logging.Init(config.TelemetryName); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	requests.RegisterCommonValidationMessages()
	shutdownTimeout := config.GracefulShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 * time.Second
	}

	shutdownTelemetry, err := configureTelemetry(ctx, config)
	if err != nil {
		return err
	}

	defer func() {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		// Ensure context is cancelled after shutdown.
		// Defer inside this function to ensure it runs after shutdownTelemetry completes.
		defer cancel()

		// Shutdown telemetry before exiting.
		shutdownTelemetry(ctxShutdown)
	}()

	v := validator.New()
	if config.ConfigureValidation != nil {
		if err := config.ConfigureValidation(v); err != nil {
			return fmt.Errorf("failed to register validations: %w", err)
		}
	}

	handler, shutdown, err := config.CreateHandler(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP handler: %w", err)
	}

	srv := getHttpServerConfig(config)
	srv.Handler = handler

	log.Printf("Starting server on port %s...\n", config.Port)
	srvErr := make(chan error, 1)
	go func() {
		// Send startup errors to the main goroutine so cleanup can run.
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	// Listen for shutdown signal
	ctxQuit, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var serverErr error
	select {
	case serverErr = <-srvErr:
		// Server failed to start; trigger cleanup path.
		serverErr = fmt.Errorf("HTTP server error: %w", serverErr)
	case <-ctxQuit.Done():
	}

	if serverErr == nil {
		log.Println("Shutting down server...")
	}

	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	g, ctxShutDown := errgroup.WithContext(ctxShutDown)

	g.Go(func() error {
		// Stop accepting new connections and drain existing ones.
		if err := srv.Shutdown(ctxShutDown); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		// Allow services to release external resources.
		if shutdown == nil {
			return nil
		}

		if err := shutdown(); err != nil {
			return fmt.Errorf("handler shutdown failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if serverErr != nil {
		return serverErr
	}

	log.Println("Server closed successfully")
	return nil
}

func configureTelemetry(ctx context.Context, config ServerRunConfiguration) (func(context.Context), error) {
	tr, err := telemetry.Init(ctx, config.TelemetryName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s tracing: %w", config.TelemetryName, err)
	}

	return func(shutdownCtx context.Context) {
		if err := tr.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shut down %s tracer provider: %v", config.TelemetryName, err)
		}
	}, nil
}

func getHttpServerConfig(config ServerRunConfiguration) *http.Server {
	srv := &http.Server{
		Addr:         ":" + config.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if config.Server.ReadTimeout != 0 {
		srv.ReadTimeout = config.Server.ReadTimeout
	}

	if config.Server.WriteTimeout != 0 {
		srv.WriteTimeout = config.Server.WriteTimeout
	}

	if config.Server.IdleTimeout != 0 {
		srv.IdleTimeout = config.Server.IdleTimeout
	}

	return srv
}
