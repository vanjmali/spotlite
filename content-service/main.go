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

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/common-lib/validations"
	"github.com/vanjmali/spotlite/content/handlers"
	"github.com/vanjmali/spotlite/content/infrastructure/mongo"
	"github.com/vanjmali/spotlite/content/repositories"
	"github.com/vanjmali/spotlite/content/routers"
	"github.com/vanjmali/spotlite/content/services"
)

var port = utils.GetEnv("APP_PORT", "3000")

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Initialize telemetry
	tr, err := telemetry.Init(ctx, "content-service")
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	defer func() {
		if err := tr.Shutdown(ctx); err != nil {
			log.Printf("failed to shut down tracer provider: %v", err)
		}
	}()

	// Initialize clients
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return fmt.Errorf("cannot start application without DB connection: %w", err)
	}

	// Configure validators
	requests.RegisterCommonValidationMessages()
	v := validator.New()
	if err := requests.RegisterValidation(v, validations.CheckValidName); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	defer dbc.Disconnect(context.Background())

	// Repository initialization
	ar := repositories.NewArtistRepository(mongo.DatabaseName(), "artists", dbc)
	sr := repositories.NewSongRepository(mongo.DatabaseName(), "songs", dbc)
	alr := repositories.NewAlbumRepository(mongo.DatabaseName(), "albums", dbc)

	// Services initialization
	as := services.NewArtistService(*ar)
	ss := services.NewSongService(*sr, *as)
	als := services.NewAlbumService(*alr, *as, *ss)

	// Handlers initialization
	ah := handlers.NewArtistHandler(*as, *v)
	sh := handlers.NewSongHandler(*ss, *v)
	alh := handlers.NewAlbumHandler(*als, *v)

	r := routers.HandleRequests(ah, sh, alh)

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

	// starts  the http server in a new goroutine so graceful shutdown mechanism doesn't get blocked and can
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
