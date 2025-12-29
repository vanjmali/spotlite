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
	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mailing"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/internal/asynqinfra"
	"github.com/vanjmali/spotlite/user-service/internal/worker"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
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
	tr, err := telemetry.Init(ctx, "user-service")
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

	mc, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = dbc.Disconnect(context.Background())
		return fmt.Errorf("cannot start application without mailing service: %w", err)
	}

	// Utility function which seeds the database with users so we could test out the email scheduler
	// load.TestLoadSeed(dbc)

	// Configure validators
	requests.RegisterCommonValidationMessages()
	v := validator.New()

	if err := requests.RegisterValidation(v, validation.CheckStrongPassword); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	if err := requests.RegisterValidation(v, validation.CheckValidUsername); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	if err := requests.RegisterValidation(v, validation.CheckValidName); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	defer dbc.Disconnect(context.Background())
	defer mc.Close()

	// repository initialization
	ur := repositories.NewRepository(mongo.DatabaseName(), "users", dbc)
	rtr := repositories.NewRefreshTokenRepository(mongo.DatabaseName(), repositories.RefreshTokensColl, dbc)
	if err := rtr.EnsureRefreshIndexes(context.Background()); err != nil {
		return fmt.Errorf("failed to ensure refresh token indexes: %w", err)
	}

	// service initialization
	ms := services.InitMailingService(mc)
	us := services.NewUserService(*ur, *ms)
	rts := services.NewRefreshTokenService(*rtr)

	redAddr := utils.MustGetEnv("REDIS_ADDR")
	redConn := asynq.RedisClientOpt{Addr: redAddr}

	// Initialize asynq service, which will initialize an asynq server, client, scheduler
	as := asynqinfra.New(redConn, 10)

	// Initialize user worker which is in charge of handling tasks
	userWorker := worker.NewUserWorker(
		as.Client(),
		us,
		ms,
	)

	// Initialize Task router which will map tasks with adequate workers
	mux := asynqinfra.NewTaskRouter(userWorker)

	// Initialize a scheduler
	//    minutes *    hours *    day of month *     month *    day of week *
	as.RegisterSchedule("53 16 * * *")

	// Starts task router and scheduler in separate go routines
	as.Start(mux)

	uh := handlers.NewUserHandler(*us, *v, *rts)
	rth := handlers.NewRefreshTokenHandler(*rts, *us, *v)

	r := routers.HandleRequests(uh, rth)

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

	// Making sure we stop the scheduler, server, client
	err = as.Stop()
	if err != nil {
		return err
	}

	log.Println("DEBUG: Shutdown complete")

	return nil
}
