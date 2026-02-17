package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/notification-service/consumers"
	"github.com/vanjmali/spotlite/notification-service/handlers"
	"github.com/vanjmali/spotlite/notification-service/infrastructure"
	"github.com/vanjmali/spotlite/notification-service/repositories"
	"github.com/vanjmali/spotlite/notification-service/routers"
	"github.com/vanjmali/spotlite/notification-service/services"
)

var (
	cassHost     = utils.GetEnv("CASSANDRA_HOST", "127.0.0.1")
	ks           = utils.GetEnv("CASSANDRA_KEYSPACE", "notification_service")
	certFilePath = utils.MustGetEnv("CERT_PATH")
	keyFilePath  = utils.MustGetEnv("KEY_PATH")
	config       = server.ServerRunConfiguration{
		TelemetryName: "notification-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			cs, jsc, rc, err := createClients(ctx)
			if err != nil {
				err = fmt.Errorf("failed to create clients: %w", err)
				return h, shutdown, err
			}

			defer func() {
				if err == nil {
					return
				}

				jsc.Close()
				_ = rc.Close()
				cs.Close()
			}()

			b := infrastructure.NewBroker()

			err = jsc.EnsureStream(ctx, events.SUBSCRIPTIONS_STREAM, []string{events.SUBJECT_SUBSCRIBER_BATCH})
			if err != nil {
				err = fmt.Errorf("failed to ensure NATS stream: %w", err)
				return h, shutdown, err
			}

			nr := createRepositories(cs)
			ns := createServices(nr, rc, b)
			h = createHandlers(ns, b)
			c := createConsumers(ns)

			consumerCtx, consumerCancel := context.WithCancel(ctx)
			consumerDone := make(chan struct{})
			var consumerErr error

			go func() {
				consumerErr = jsc.StartConsumer(
					consumerCtx,
					events.SUBSCRIPTIONS_STREAM,
					events.SUBJECT_SUBSCRIBER_BATCH,
					events.SUB_DURABLE,
					c.HandleSubscribersBatch,
				)
				close(consumerDone)
			}()

			go b.Listen(consumerCtx)

			go func() {
				<-consumerDone
				if consumerErr != nil && !errors.Is(consumerErr, context.Canceled) {
					log.Printf("notification consumer stopped unexpectedly: %v", consumerErr)
				}
			}()

			shutdown = func() error {
				var errs []error

				consumerCancel()
				<-consumerDone

				if consumerErr != nil && !errors.Is(consumerErr, context.Canceled) {
					errs = append(errs, fmt.Errorf("subscription consumer error: %w", consumerErr))
				}

				jsc.Close()
				if err := rc.Close(); err != nil {
					errs = append(errs, fmt.Errorf("redis error: %w", err))

				}
				cs.Close()

				return errors.Join(errs...)
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

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start notification service: %v", err)
	}
}

func createClients(ctx context.Context) (*gocql.Session, *events.JetStreamClient, *redis.Client, error) {
	// schema initialization
	if err := infrastructure.InitializeSchema(cassHost, ks); err != nil {
		return nil, nil, nil, err
	}

	// initialize cassandra session which will be used to execute queries
	cs, err := infrastructure.Initialize(cassHost, ks)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize database session: %w", err)
	}

	// initialize redis client
	rc, err := infrastructure.InitRedis(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	jsc, err := events.NewClient("nats://nats:4222")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialized NATS jet stream client: %w", err)
	}

	return cs, jsc, rc, nil
}

func createHandlers(
	ns *services.NotificationService,
	b *infrastructure.Broker,
) http.Handler {
	nh := handlers.NewNotificationHandler(ns, b)
	return routers.HandleRequests(nh)
}

func createServices(
	nr *repositories.NotificationRepository,
	rc *redis.Client,
	b *infrastructure.Broker,
) *services.NotificationService {
	ns := services.NewNotificationService(nr, rc, b)
	return ns
}

func createRepositories(cs *gocql.Session) *repositories.NotificationRepository {
	nr := repositories.NewNotificationRepository(cs)
	return nr
}

func createConsumers(ns *services.NotificationService) *consumers.NotificationConsumer {
	sc := consumers.NewConsumer(ns)
	return sc
}
