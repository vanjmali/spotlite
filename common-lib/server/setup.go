package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
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
		ReadTimeout  time.Duration
		// WriteTimeout is the maximum duration before timing out writes of the response. Default is 15 seconds.
		WriteTimeout time.Duration
		// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled. Default is 60 seconds.
		IdleTimeout  time.Duration
	}
}

// Run starts the HTTP server based on the provided configuration.
// It handles graceful shutdown on receiving termination signals.
func Run(ctx context.Context, config ServerRunConfiguration) error {
	requests.RegisterCommonValidationMessages()
	shutdownTelemetry, err := configureTelemetry(ctx, config)
	if err != nil {
		return err
	}

	defer shutdownTelemetry()

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
	ctxQuit, cancel := context.WithCancel(ctx)
	go listenForShutdownSignal(cancel)

	var serverErr error
	select {
	case serverErr = <-srvErr:
		// Server failed to start; trigger cleanup path.
		serverErr = fmt.Errorf("HTTP server error: %w", serverErr)
		cancel()
	case <-ctxQuit.Done():
	}

	if serverErr == nil {
		log.Println("Shutting down server...")
	}

	shutdownTimeout := config.GracefulShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 * time.Second
	}

	ctxShutDown, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	g, ctxShutDown := errgroup.WithContext(ctxShutDown)

	g.Go(func() error {
		// Stop accepting new connections and drain existing ones.
		if err := srv.Shutdown(ctxShutDown); err != nil {
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

func configureTelemetry(ctx context.Context, config ServerRunConfiguration) (func(), error) {
	tr, err := telemetry.Init(ctx, config.TelemetryName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s tracing: %w", config.TelemetryName, err)
	}

	return func() {
		if err := tr.Shutdown(ctx); err != nil {
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

func listenForShutdownSignal(cancelFunc context.CancelFunc) {
	// Listen for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, exiting...")
	cancelFunc()
}
