package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vanjmali/spotlite/analytics-service/consumers"
	"github.com/vanjmali/spotlite/analytics-service/handlers"
	"github.com/vanjmali/spotlite/analytics-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/analytics-service/repositories"
	"github.com/vanjmali/spotlite/analytics-service/routers"
	"github.com/vanjmali/spotlite/analytics-service/services"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

var (
	rootCACertFilePath = utils.MustGetEnv("ROOT_CERT_PATH")
	certFilePath       = utils.MustGetEnv("CERT_PATH")
	keyFilePath        = utils.MustGetEnv("KEY_PATH")
	config             = server.ServerRunConfiguration{
		TelemetryName: "analytics-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		ConfigureValidation: func(v *validator.Validate) error {
			return nil
		},
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			mc, jsc, err := createClients()
			if err != nil {
				return h, shutdown, fmt.Errorf("failed to create clients: %w", err)
			}

			// Cleanup resources on error
			defer func() {
				if err == nil {
					return
				}
				_ = mc.Disconnect(ctx)
				jsc.Close()
			}()

			// Initialize event store indexes
			err = initializeEventStoreIndexes(ctx, mc)
			if err != nil {
				err = fmt.Errorf("failed to initialize event store indexes: %w", err)
				return h, shutdown, err
			}

			// Initialize read model indexes
			err = initializeReadModelIndexes(ctx, mc)
			if err != nil {
				err = fmt.Errorf("failed to initialize read model indexes: %w", err)
				return h, shutdown, err
			}

			// Create shared analytics service for both HTTP handlers and NATS consumers
			analyticsService := createAnalyticsService(mc)

			// Create HTTP handlers
			ah := handlers.NewAnalyticsHandler(analyticsService, v)
			h = routers.HandleRequests(ah)

			// Create NATS consumers
			c := consumers.NewConsumer(analyticsService)

			// Setup consumer goroutines
			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup

			// Configure all analytics event consumers
			configs := []events.ConsumerConfig{
				{
					Stream:  events.LISTENS_STREAM,
					Subject: events.SUBJECT_LISTEN_CREATED,
					Durable: events.LISTEN_DURABLE,
					Handler: c.HandleListenCreated,
				},
				{
					Stream:  events.RATINGS_STREAM,
					Subject: events.SUBJECT_RATING_CREATED,
					Durable: events.RATING_DURABLE,
					Handler: c.HandleRatingCreated,
				},
				{
					Stream:  events.RATINGS_STREAM,
					Subject: events.SUBJECT_RATING_UPDATED,
					Durable: events.RATING_DURABLE,
					Handler: c.HandleRatingUpdated,
				},
				{
					Stream:  events.RATINGS_STREAM,
					Subject: events.SUBJECT_RATING_DELETED,
					Durable: events.RATING_DURABLE,
					Handler: c.HandleRatingDeleted,
				},
				{
					Stream:  events.SUBSCRIPTIONS_STREAM,
					Subject: events.SUBJECT_SUBSCRIPTION_CREATED,
					Durable: events.SUBSCRIPTION_DURABLE,
					Handler: c.HandleSubscriptionCreated,
				},
				{
					Stream:  events.SUBSCRIPTIONS_STREAM,
					Subject: events.SUBJECT_SUBSCRIPTION_DELETED,
					Durable: events.SUBSCRIPTION_DURABLE,
					Handler: c.HandleSubscriptionDeleted,
				},
			}

			consumerErrs := make([]error, len(configs))
			for i, cfg := range configs {
				consumerWg.Add(1)

				// pass 'i' and 'cfg' into the goroutine to capture them correctly
				go func(index int, config events.ConsumerConfig) {
					defer consumerWg.Done()
					err := jsc.StartConsumer(consumerCtx, config.Stream, config.Subject, config.Durable, config.Handler)
					if err != nil {
						logging.Errorf(ctx, "consumer %s stopped with error: %v", config.Durable, err)
						consumerErrs[index] = err
					}
				}(i, cfg)
			}

			shutdown = func() error {
				var errs []error

				// Signal all consumers to stop
				consumerCancel()
				consumerWg.Wait()

				// Check if any consumer encountered a critical error
				for _, consumerErr := range consumerErrs {
					if consumerErr != nil && !errors.Is(consumerErr, context.Canceled) {
						errs = append(errs, consumerErr)
					}
				}

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Disconnect MongoDB
				if err := mc.Disconnect(shutdownCtx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
					errs = append(errs, fmt.Errorf("mongo disconnect error: %w", err))
				}

				// Close NATS connection
				jsc.Close()

				return errors.Join(errs...)
			}

			return h, shutdown, nil
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

func createClients() (*mongodriver.Client, *events.JetStreamClient, error) {
	mc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	jsc, err := events.NewClient("tls://nats:4222", nats.RootCAs(rootCACertFilePath))
	if err != nil {
		_ = mc.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("failed to initialize NATS JetStream client: %w", err)
	}

	return mc, jsc, nil
}

func initializeEventStoreIndexes(ctx context.Context, mongoClient *mongodriver.Client) error {
	dbName := utils.MustGetEnv("DB_NAME")
	esr := repositories.NewEventStoreRepository(dbName, "events", mongoClient)
	return esr.EnsureIndexes(ctx)
}

func initializeReadModelIndexes(ctx context.Context, mongoClient *mongodriver.Client) error {
	dbName := utils.MustGetEnv("DB_NAME")

	// Initialize user analytics read model indexes
	uar := repositories.NewUserAnalyticsRepository(dbName, "user_analytics", mongoClient)
	if err := uar.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("failed to initialize user analytics indexes: %w", err)
	}

	// Initialize user activity history read model indexes
	uahr := repositories.NewUserActivityHistoryRepository(dbName, "user_activity_history", mongoClient)
	if err := uahr.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("failed to initialize user activity history indexes: %w", err)
	}

	return nil
}

// createAnalyticsService creates a shared analytics service instance
// Used by both HTTP handlers and NATS event consumers
func createAnalyticsService(mongoClient *mongodriver.Client) *services.AnalyticsService {
	dbName := utils.MustGetEnv("DB_NAME")

	// Create repositories
	eventStoreRepo := repositories.NewEventStoreRepository(dbName, "events", mongoClient)
	analyticsRepo := repositories.NewUserAnalyticsRepository(dbName, "user_analytics", mongoClient)
	historyRepo := repositories.NewUserActivityHistoryRepository(dbName, "user_activity_history", mongoClient)

	// Create and return analytics service
	return services.NewAnalyticsService(eventStoreRepo, analyticsRepo, historyRepo)
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		logging.Errorf(context.Background(), "failed to start analytics service: %v", err)
	}
}
