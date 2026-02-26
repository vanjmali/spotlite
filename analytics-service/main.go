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
	"github.com/vanjmali/spotlite/analytics-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/analytics-service/repositories"
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

			// Initialize NATS stream for analytics events
			err = jsc.EnsureStream(
				ctx,
				events.ANALYTICS_STREAM,
				[]string{
					events.SUBJECT_SONG_PLAYED,
					events.SUBJECT_RATING_CREATED,
					events.SUBJECT_RATING_UPDATED,
					events.SUBJECT_RATING_DELETED,
					events.SUBJECT_SUBSCRIPTION_CREATED,
					events.SUBJECT_SUBSCRIPTION_DELETED,
					events.SUBJECT_SONG_DELETED,
				},
			)
			if err != nil {
				err = fmt.Errorf("failed to ensure NATS stream: %w", err)
				return h, shutdown, err
			}

			c := createConsumers()

			// Setup consumer goroutines
			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup

			// Configure all analytics event consumers
			configs := []events.ConsumerConfig{
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_SONG_PLAYED,
					Durable: events.SONG_PLAYED_DURABLE,
					Handler: c.HandleSongPlayed,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_RATING_CREATED,
					Durable: events.RATING_CREATED_DURABLE,
					Handler: c.HandleRatingCreated,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_RATING_UPDATED,
					Durable: events.RATING_UPDATED_DURABLE,
					Handler: c.HandleRatingUpdated,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_RATING_DELETED,
					Durable: events.RATING_DELETED_DURABLE,
					Handler: c.HandleRatingDeleted,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_SUBSCRIPTION_CREATED,
					Durable: events.SUBSCRIPTION_CREATED_DURABLE,
					Handler: c.HandleSubscriptionCreated,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_SUBSCRIPTION_DELETED,
					Durable: events.SUBSCRIPTION_DELETED_DURABLE,
					Handler: c.HandleSubscriptionDeleted,
				},
				{
					Stream:  events.ANALYTICS_STREAM,
					Subject: events.SUBJECT_SONG_DELETED,
					Durable: events.SONG_DELETED_DURABLE,
					Handler: c.HandleSongDeleted,
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

func createConsumers() *consumers.AnalyticsConsumer {
	return consumers.NewConsumer()
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		logging.Errorf(context.Background(), "failed to start analytics service: %v", err)
	}
}
