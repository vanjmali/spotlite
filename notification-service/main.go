package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
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
	cassHost = utils.GetEnv("CASSANDRA_HOST", "127.0.0.1")
	ks       = utils.GetEnv("CASSANDRA_KEYSPACE", "notification_service")
)

var config = server.ServerRunConfiguration{
	TelemetryName: "notification-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		cs, jsc, err := createClients()
		if err != nil {
			err = fmt.Errorf("failed to create clients: %w", err)
			return h, shutdown, err
		}

		b := infrastructure.NewBroker()
		go b.Listen()

		_ = jsc.EnsureStream(ctx, events.SUBSCRIPTIONS_STREAM, []string{events.SUBJECT_SUBSCRIBER_BATCH})

		nr := createRepositories(cs)
		ns := createServices(nr, b)
		h = createHandlers(ns, b)
		c := createConsumers(ns)

		go jsc.StartConsumer(ctx, events.SUBSCRIPTIONS_STREAM, events.SUBJECT_SUBSCRIBER_BATCH, events.SUB_DURABLE, c.HandleSubscribersBatch)

		shutdown = func() error {
			cs.Close()
			return nil
		}

		return h, shutdown, err
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start notification service: %v", err)
	}
}

func createClients() (*gocql.Session, *events.JetStreamClient, error) {
	// schema initialization
	if err := infrastructure.InitializeSchema(cassHost, ks); err != nil {
		return nil, nil, err
	}

	// initialize cassandra session which will be used to execute queries
	cs, err := infrastructure.Initialize(cassHost, ks)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database session: %w", err)
	}

	jsc, err := events.NewClient("nats://nats:4222")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialized NATS jet stream client: %w", err)
	}

	return cs, jsc, nil
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
	b *infrastructure.Broker,
) *services.NotificationService {
	ns := services.NewNotificationService(nr, b)
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
