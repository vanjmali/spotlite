package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/common-lib/validations"
	"github.com/vanjmali/spotlite/user-service/handlers"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mailing"
	"github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/user-service/internal/asynqinfra"
	"github.com/vanjmali/spotlite/user-service/internal/worker"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/routers"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
	"github.com/wneessen/go-mail"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "user-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		requests.RegisterJSONTagNameFunc(v)
		if err := requests.RegisterValidation(v, validation.CheckStrongPassword); err != nil {
			return fmt.Errorf("failed to register strong password validation: %w", err)
		}

		if err := requests.RegisterValidation(v, validation.CheckValidUsername); err != nil {
			return fmt.Errorf("failed to register username validation: %w", err)
		}

		if err := requests.RegisterValidation(v, validations.CheckValidName); err != nil {
			return fmt.Errorf("failed to register name validation: %w", err)
		}

		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		mc, mail, err := createClients(ctx)
		if err != nil {
			err = fmt.Errorf("failed to create clients: %w", err)
			return h, shutdown, err
		}
		// Cleanup resources on error
		var asynqShutdown func() error
		defer func() {
			if err == nil {
				// No error, do nothing when function exits
				return
			}
			_ = mc.Disconnect(ctx)
			_ = mail.Close()
			if asynqShutdown != nil {
				_ = asynqShutdown()
			}
		}()

		ur, rtr, prr, err := createRepositories(ctx, mc)
		if err != nil {
			err = fmt.Errorf("failed to create repositories: %w", err)
			return h, shutdown, err
		}

		ms, us, rts, prs := createServices(mail, ur, rtr, prr)
		h = createHandlers(v, us, rts, prs)

		asynqShutdown = setupAsynq(us, ms)
		shutdown = func() error {
			if err := mc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}

			if err := mail.Close(); err != nil {
				return fmt.Errorf("failed to close mail client: %w", err)
			}

			if err := asynqShutdown(); err != nil {
				return fmt.Errorf("failed to shutdown asynq: %w", err)
			}

			return nil
		}

		return h, shutdown, err
	},
	Server: struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
		CertFilePath string
		KeyFilePath  string
	}{
		CertFilePath: utils.MustGetEnv("CERT_PATH"),
		KeyFilePath:  utils.MustGetEnv("KEY_PATH"),
	},
}

func createClients(ctx context.Context) (*mongodriver.Client, *mail.Client, error) {
	mongo, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	mail, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = mongo.Disconnect(ctx)
		return nil, nil, fmt.Errorf("failed to initialize mail client: %w", err)
	}

	return mongo, mail, nil
}

func createRepositories(ctx context.Context, mongo *mongodriver.Client) (
	services.UserRepository,
	services.RefreshTokenRepository,
	services.PasswordRecoveryRepository,
	error,
) {
	name := utils.MustGetEnv("DB_NAME")
	ur := repositories.NewUserRepositoryMongo(name, "users", mongo)
	rtr := repositories.NewRefreshTokenRepository(name, "refresh_tokens", mongo)
	prr := repositories.NewPasswordRecoveryRepository(name, "password_recovery_tokens", mongo)

	if err := ur.EnsureUserIndexes(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to ensure user indexes: %w", err)
	}

	if err := rtr.EnsureRefreshIndexes(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to ensure refresh token indexes: %w", err)
	}

	return ur, rtr, prr, nil
}

func createServices(
	mail *mail.Client,
	ur services.UserRepository,
	rr services.RefreshTokenRepository,
	pt services.PasswordRecoveryRepository) (
	*services.MailService,
	*services.UserService,
	*services.RefreshTokenService,
	*services.PasswordRecoveryService,
) {
	mailCfg := services.MailConfig{
		VerificationURL:  utils.MustGetEnv("SRV_USER_VERIFY_URL"),
		PasswordResetURL: utils.MustGetEnv("SRV_USER_PASSWORD_RESET_URL"),
		MailFromAddress:  utils.MustGetEnv("MAIL_FROM"),
	}

	ms := services.InitMailingService(mail, mailCfg)
	us := services.NewUserService(ur, ms)
	rts := services.NewRefreshTokenService(rr)
	prs := services.NewPasswordRecoveryService(ur, pt, ms)

	return ms, us, rts, prs
}

func createHandlers(
	v *validator.Validate,
	us *services.UserService,
	rts *services.RefreshTokenService,
	prs *services.PasswordRecoveryService,
) http.Handler {
	uh := handlers.NewUserHandler(*us, *v, *rts)

	rth := handlers.NewRefreshTokenHandler(*rts, *us, *v)
	prh := handlers.NewPasswordRecoveryHandler(*prs, *v)

	return routers.HandleRequests(uh, rth, prh)
}

// setupAsynq initializes and starts the asynq server, client, scheduler, workers and task router.
// Returns a shutdown function to gracefully stop asynq components.
func setupAsynq(us *services.UserService, ms *services.MailService) func() error {
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
	as.RegisterSchedule("01 19 * * *")

	// Starts task router and scheduler in separate go routines
	as.Start(mux)

	return as.Stop
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start user service: %v", err)
	}
}
