package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mailing"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/internal/tasks"
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
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return fmt.Errorf("cannot start application without DB connection: %w", err)
	}

	mc, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = dbc.Disconnect(context.Background())
		return fmt.Errorf("cannot start application without mailing service: %w", err)
	}

	v := validator.New()
	if err := requests.RegisterValidation(v, validation.CheckStrongPassword); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	if err := requests.RegisterValidation(v, validation.CheckValidUsername); err != nil {
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

	redAddr := os.Getenv("REDIS_ADDR")
	if redAddr == "" {
		redAddr = "127.0.0.1:6379"
	}
	redConn := asynq.RedisClientOpt{Addr: redAddr}

	// asynq server initialization, most 10 tasks will be handled in parallel
	as := asynq.NewServer(redConn, asynq.Config{Concurrency: 10})

	// asynq client initialization
	ac := asynq.NewClient(redConn)
	defer ac.Close()

	// worker initialization
	w := worker.NewUserWorker(ac, us, ms)

	// worker router initialization
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypePasswordExpiryCheck, w.HandleExpiryCheck2)
	mux.HandleFunc(tasks.TypeSendExpiryEmail, w.HandleSendExpiryEmail)

	// Starting Asynq server in a new goroutine, which will act as a background worker and waits
	// for new tasks to be added to Redis (Redis is being used as a message broker in this scenario).
	// AKA Consumer,
	go as.Run(mux)

	sch := asynq.NewScheduler(redConn, nil)
	if _, err := sch.Register("17 18 * * *", tasks.NewPasswordExpiryCheckTask()); err != nil {
		log.Fatal(err)
	}

	// Starting the Scheduler in another goroutine, which will check for the schedule we defined
	// and when the time comes push the task to message broker,
	// AKA Producer,
	go sch.Run()

	uh := handlers.NewUserHandler(*us, *v, *rts)
	rth := handlers.NewRefreshTokenHandler(*rts, *us, *v)

	r := routers.HandleRequests(uh, rth)

	srvAddr := ":" + port

	log.Printf("Listening on %s", srvAddr)
	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
