package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/common-lib/validations"
	"github.com/vanjmali/spotlite/user-service/consumers"
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

var (
	userListenDurable              = "USER_LISTEN_PROCESSOR"
	userRatingCreatedDurable       = "USER_RATING_CREATED_PROCESSOR"
	userRatingUpdatedDurable       = "USER_RATING_UPDATED_PROCESSOR"
	userSubscriptionCreatedDurable = "USER_SUBSCRIPTION_CREATED_PROCESSOR"
	userSubscriptionDeletedDurable = "USER_SUBSCRIPTION_DELETED_PROCESSOR"
	rootCACertFilePath             = utils.MustGetEnv("ROOT_CERT_PATH")
	certFilePath                   = utils.MustGetEnv("CERT_PATH")
	keyFilePath                    = utils.MustGetEnv("KEY_PATH")
	natsURL                        = utils.MustGetEnv("NATS_URL")
	config                         = server.ServerRunConfiguration{
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
			mc, mail, jsc, err := createClients(ctx)
			if err != nil {
				err = fmt.Errorf("failed to create clients: %w", err)
				return h, shutdown, err
			}
			// Cleanup resources on error
			var asynqShutdown func() error
			var stopConsumers func()
			var consumerWg sync.WaitGroup
			defer func() {
				if err == nil {
					// No error, do nothing when function exits
					return
				}

				if stopConsumers != nil {
					stopConsumers()
					consumerWg.Wait()
				}

				_ = mc.Disconnect(ctx)
				_ = mail.Close()
				jsc.Close()

				if asynqShutdown != nil {
					_ = asynqShutdown()
				}
			}()

			err = jsc.EnsureStream(ctx, events.USERS_STREAM, []string{events.SUBJECT_USER_CREATED})
			if err != nil {
				err = fmt.Errorf("failed to ensure song stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(
				ctx,
				events.LISTENS_STREAM,
				[]string{events.SUBJECT_LISTEN_CREATED},
			); err != nil {
				err = fmt.Errorf("failed to ensure listens stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(
				ctx,
				events.RATINGS_STREAM,
				[]string{
					events.SUBJECT_RATING_CREATED,
					events.SUBJECT_RATING_UPDATED,
				},
			); err != nil {
				err = fmt.Errorf("failed to ensure ratings stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(
				ctx,
				events.SUBSCRIPTIONS_STREAM,
				[]string{events.SUBJECT_SUBSCRIPTION_CREATED, events.SUBJECT_SUBSCRIPTION_DELETED},
			); err != nil {
				err = fmt.Errorf("failed to ensure subscriptions stream: %w", err)
				return h, shutdown, err
			}

			ur, rtr, prr, uar, err := createRepositories(ctx, mc)
			if err != nil {
				err = fmt.Errorf("failed to create repositories: %w", err)
				return h, shutdown, err
			}

			ms, us, rts, prs, uas := createServices(mail, ur, rtr, prr, uar, jsc)
			h = createHandlers(v, us, rts, prs, uas)

			stopConsumers = setupActivityConsumers(ctx, jsc, uas, &consumerWg)

			asynqShutdown = setupAsynq(us, ms)
			shutdown = func() error {
				if stopConsumers != nil {
					stopConsumers()
					consumerWg.Wait()
				}

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
			CertFilePath: certFilePath,
			KeyFilePath:  keyFilePath,
		},
	}
)

func createClients(ctx context.Context) (*mongodriver.Client, *mail.Client, *events.JetStreamClient, error) {
	mongo, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	mail, err := mailing.InitClientFromEnv()
	if err != nil {
		_ = mongo.Disconnect(ctx)
		return nil, nil, nil, fmt.Errorf("failed to initialize mail client: %w", err)
	}

	jsc, err := events.NewClient(natsURL, nats.RootCAs(rootCACertFilePath))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialized NATS jet strea, client: %w", err)
	}

	return mongo, mail, jsc, nil
}

func createRepositories(ctx context.Context, mongo *mongodriver.Client) (
	services.UserRepository,
	services.RefreshTokenRepository,
	services.PasswordRecoveryRepository,
	services.UserActivityRepository,
	error,
) {
	name := utils.MustGetEnv("DB_NAME")
	ur := repositories.NewUserRepositoryMongo(name, "users", mongo)
	rtr := repositories.NewRefreshTokenRepository(name, "refresh_tokens", mongo)
	prr := repositories.NewPasswordRecoveryRepository(name, "password_recovery_tokens", mongo)
	uar := repositories.NewUserActivityRepository(name, "user_activity_events", mongo)

	if err := ur.EnsureUserIndexes(ctx); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to ensure user indexes: %w", err)
	}

	if err := rtr.EnsureRefreshIndexes(ctx); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to ensure refresh token indexes: %w", err)
	}

	if err := uar.EnsureIndexes(ctx); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to ensure user activity indexes: %w", err)
	}

	return ur, rtr, prr, uar, nil
}

func createServices(
	mail *mail.Client,
	ur services.UserRepository,
	rr services.RefreshTokenRepository,
	pt services.PasswordRecoveryRepository,
	uar services.UserActivityRepository,
	jsc *events.JetStreamClient) (
	*services.MailService,
	*services.UserService,
	*services.RefreshTokenService,
	*services.PasswordRecoveryService,
	*services.UserActivityService,
) {
	mailCfg := services.MailConfig{
		VerificationURL:  utils.MustGetEnv("SRV_USER_VERIFY_URL"),
		PasswordResetURL: utils.MustGetEnv("SRV_USER_PASSWORD_RESET_URL"),
		MailFromAddress:  utils.MustGetEnv("MAIL_FROM"),
	}

	ms := services.InitMailingService(mail, mailCfg)
	us := services.NewUserService(ur, ms, jsc)
	rts := services.NewRefreshTokenService(rr)
	prs := services.NewPasswordRecoveryService(ur, pt, ms)
	uas := services.NewUserActivityService(uar)

	return ms, us, rts, prs, uas
}

func createHandlers(
	v *validator.Validate,
	us *services.UserService,
	rts *services.RefreshTokenService,
	prs *services.PasswordRecoveryService,
	uas *services.UserActivityService,
) http.Handler {
	uh := handlers.NewUserHandler(*us, *v, *rts)
	ah := handlers.NewUserActivityHandler(uas)

	rth := handlers.NewRefreshTokenHandler(*rts, *us, *v)
	prh := handlers.NewPasswordRecoveryHandler(*prs, *v)

	return routers.HandleRequests(uh, rth, prh, ah)
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

func setupActivityConsumers(
	ctx context.Context,
	jsc *events.JetStreamClient,
	uas *services.UserActivityService,
	wg *sync.WaitGroup,
) func() {
	consumer := consumers.NewUserActivityConsumer(uas)
	consumerCtx, cancel := context.WithCancel(ctx)

	configs := []events.ConsumerConfig{
		{
			Stream:  events.LISTENS_STREAM,
			Subject: events.SUBJECT_LISTEN_CREATED,
			Durable: userListenDurable,
			Handler: consumer.HandleListenCreated,
		},
		{
			Stream:  events.RATINGS_STREAM,
			Subject: events.SUBJECT_RATING_CREATED,
			Durable: userRatingCreatedDurable,
			Handler: consumer.HandleRatingCreated,
		},
		{
			Stream:  events.RATINGS_STREAM,
			Subject: events.SUBJECT_RATING_UPDATED,
			Durable: userRatingUpdatedDurable,
			Handler: consumer.HandleRatingUpdated,
		},
		{
			Stream:  events.SUBSCRIPTIONS_STREAM,
			Subject: events.SUBJECT_SUBSCRIPTION_CREATED,
			Durable: userSubscriptionCreatedDurable,
			Handler: consumer.HandleSubscriptionCreated,
		},
		{
			Stream:  events.SUBSCRIPTIONS_STREAM,
			Subject: events.SUBJECT_SUBSCRIPTION_DELETED,
			Durable: userSubscriptionDeletedDurable,
			Handler: consumer.HandleSubscriptionDeleted,
		},
	}

	for _, cfg := range configs {
		wg.Add(1)
		go func(config events.ConsumerConfig) {
			defer wg.Done()
			if err := jsc.StartConsumer(
				consumerCtx,
				config.Stream,
				config.Subject,
				config.Durable,
				config.Handler,
			); err != nil && !errors.Is(err, context.Canceled) {
				logging.Errorf(context.Background(), "activity consumer %s failed: %v", config.Durable, err)
			}
		}(cfg)
	}

	return cancel
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		logging.Errorf(context.Background(), "failed to start user service: %v", err)
	}
}
