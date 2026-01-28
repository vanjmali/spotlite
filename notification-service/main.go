package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/notifications/handlers"
	"github.com/vanjmali/spotlite/notifications/infrastructure"
	"github.com/vanjmali/spotlite/notifications/repositories"
	"github.com/vanjmali/spotlite/notifications/routers"
	"github.com/vanjmali/spotlite/notifications/services"
)

var (
	cassHost = utils.GetEnv("CASSANDRA_HOST", "127.0.0.1")
	ks       = utils.GetEnv("CASSANDRA_KEYSPACE", "notification_service")
)

var config = server.ServerRunConfiguration{
	TelemetryName: "notification-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		cs, err := createClients()
		if err != nil {
			err = fmt.Errorf("failed to create clients: %w", err)
			return h, shutdown, err
		}

		b := infrastructure.NewBroker()
		go b.Listen()

		nr := createRepositories(cs)
		ns := createServices(nr)
		h = createHandlers(ns, b)

		shutdown = func() error {
			cs.Close()
			return nil
		}

		return h, shutdown, err
	},
	GracefulShutdownTimeout: 10,
	Server: struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start notification service: %v", err)
	}
}

func createClients() (*gocql.Session, error) {
	// schema initialization
	if err := infrastructure.InitializeSchema(cassHost, ks); err != nil {
		return nil, err
	}

	// initialize cassandra session which will be used to execute queries
	cs, err := infrastructure.Initialize(cassHost, ks)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database session: %w", err)
	}

	return cs, nil
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
) *services.NotificationService {
	ns := services.NewNotificationService(nr)
	return ns
}

func createRepositories(cs *gocql.Session) *repositories.NotificationRepository {
	nr := repositories.NewNotificationRepository(cs)
	return nr
}
