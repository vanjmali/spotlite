package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/subscription-service/consumers"
	"github.com/vanjmali/spotlite/subscription-service/handlers"
	adapters "github.com/vanjmali/spotlite/subscription-service/infrastructure/grpc"
	"github.com/vanjmali/spotlite/subscription-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
	"github.com/vanjmali/spotlite/subscription-service/routers"
	"github.com/vanjmali/spotlite/subscription-service/services"
	"github.com/vanjmali/spotlite/subscription-service/validation"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rootCACertFilePath = utils.MustGetEnv("ROOT_CERT_PATH")
	certFilePath       = utils.MustGetEnv("CERT_PATH")
	keyFilePath        = utils.MustGetEnv("KEY_PATH")
	config             = server.ServerRunConfiguration{
		TelemetryName: "subscription-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		ConfigureValidation: func(v *validator.Validate) error {
			requests.RegisterJSONTagNameFunc(v)
			if err := requests.RegisterValidation(v, validation.CheckValidEntityID); err != nil {
				return fmt.Errorf("failed to register entity ID validation: %w", err)
			}

			if err := requests.RegisterValidation(v, validation.CheckValidSubscriptionType); err != nil {
				return fmt.Errorf("failed to register subscription type validation: %w", err)
			}

			return nil
		},
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			dbc, gc, jsc, err := createClients()
			if err != nil {
				err = fmt.Errorf("failed to create clients: %w", err)
				return h, shutdown, err
			}

			// Cleanup resources on error
			defer func() {
				if err == nil {
					return
				}

				_ = gc.Close()
				_ = dbc.Disconnect(ctx)
				jsc.Close()
			}()

			err = initializeSubscriptionIndexes(ctx, dbc)
			if err != nil {
				return nil, nil, err
			}

			if err = jsc.EnsureStream(ctx, events.CONTENT_STREAM, []string{events.SUBJECT_ENTITY_CREATED, events.SUBJECT_ENTITY_UPDATED}); err != nil {
				err = fmt.Errorf("failed to ensure content stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(ctx, events.GENRES_STREAM, []string{events.SUBJECT_GENRE_SUBSCRIBED, events.SUBJECT_GENRE_CREATED}); err != nil {
				err = fmt.Errorf("failed to ensure genres stream: %w", err)
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

			gcc := createAdapters(gc)
			sr := createRepositories(dbc)
			ss := createServices(sr, gcc, jsc)
			h = createHandlers(v, ss)
			c := createConsumers(ss)

			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup

			// configure consumers
			configs := []events.ConsumerConfig{
				{
					Stream:  events.CONTENT_STREAM,
					Subject: events.SUBJECT_ENTITY_CREATED,
					Durable: events.ENTITY_CREATE_DURABLE,
					Handler: c.HandleEntityCreated,
				},
				{
					Stream:  events.CONTENT_STREAM,
					Subject: events.SUBJECT_ENTITY_UPDATED,
					Durable: events.ENTITY_UPDATE_DURABLE,
					Handler: c.HandleEntityUpdated,
				},
			}

			consumerErrs := make([]error, len(configs))
			for i, cfg := range configs {
				consumerWg.Add(1)

				// pass 'i' and 'cfg' into the goroutine to capture them correctly
				go func(index int, config events.ConsumerConfig) {
					defer consumerWg.Done()

					err := jsc.StartConsumer(
						consumerCtx,
						config.Stream,
						config.Subject,
						config.Durable,
						config.Handler,
					)

					if err != nil && !errors.Is(err, context.Canceled) {
						logging.Errorf(context.Background(), "consumer %s stopped unexpectedly: %v", config.Durable, err)
						consumerErrs[index] = fmt.Errorf("consumer %s error: %w", config.Durable, err)
					}
				}(i, cfg)
			}

			shutdown = func() error {
				var errs []error

				consumerCancel()
				// makes sure all consumers are down before shuting down NATS stream
				consumerWg.Wait()

				for _, err := range consumerErrs {
					if err != nil {
						errs = append(errs, err)
					}
				}

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := gc.Close(); err != nil {
					errs = append(errs, fmt.Errorf("grpc close error: %w", err))
				}

				if err := dbc.Disconnect(shutdownCtx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
					errs = append(errs, fmt.Errorf("mongo disconnect error: %w", err))
				}

				// close nats jetstream connection only after the consumers are shut down, to avoid consumers communicating
				// via closed connection (potential panics)
				jsc.Close()

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
		logging.Errorf(context.Background(), "failed to start subscription service: %v", err)
	}
}

func createClients() (*mongodriver.Client, *grpc.ClientConn, *events.JetStreamClient, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize subscription service MongoDB client: %w", err)
	}

	creds, err := generateCreds()
	if err != nil {
		return nil, nil, nil, err
	}

	grpcTarget := utils.MustGetEnv("CONTENT_GRPC_ADDRESS")

	gc, err := grpc.NewClient(
		grpcTarget,
		grpc.WithTransportCredentials(creds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to establish a RPC connection with the content-service: %w", err)
	}

	jsc, err := events.NewClient("tls://nats:4222", nats.RootCAs(rootCACertFilePath))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialized NATS jets teram client: %w", err)
	}

	return dbc, gc, jsc, nil
}

func initializeSubscriptionIndexes(ctx context.Context, c *mongodriver.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models := []mongodriver.IndexModel{
		// compound index which groups and sorts subscriptions by subscriber_id and sorts it by entity_id field,
		// it provides efficient user following list lookup and has a constraint which forbids duplicate subscriptions
		//
		// other than that it also provides fast check is the user subscribed to particular content when located on
		// genre/artist page, since it will be fetched by these two fields
		{
			Keys:    bson.D{{Key: "subscriber_id", Value: 1}, {Key: "entity_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// allows efficient user following list retrieval, subscriptions are sorted in chronological order starting from
		// latest
		{
			Keys: bson.D{{Key: "subscriber_id", Value: 1}, {Key: "subscribed_at", Value: -1}},
		},
		// using this index we can efficiently fetch all subscriptions by entity_id which is useful for sending notifications
		// to users when new albums drop/ new artists of a genre are created. JSYK The compound index won't do the job.
		{
			Keys: bson.D{{Key: "entity_id", Value: 1}},
		},
	}

	_, err := c.Database(mongo.DatabaseName()).
		Collection("subscriptions").
		Indexes().
		CreateMany(ctx, models)

	return err
}

func createAdapters(gc *grpc.ClientConn) *adapters.GrpcContentEntityGetter {
	return adapters.NewGrpcContentEntityGetter(gc)
}

func createRepositories(dbc *mongodriver.Client) *repositories.SubscriptionRepository {
	name := utils.MustGetEnv("DB_NAME")
	sr := repositories.NewSubscriptionRepository(name, "subscriptions", dbc)

	return sr
}

func createServices(
	sr *repositories.SubscriptionRepository,
	gcc *adapters.GrpcContentEntityGetter,
	jsc *events.JetStreamClient,
) *services.SubscriptionService {
	ss := services.NewSubscriptionService(sr, gcc, jsc)

	return ss
}

func createHandlers(
	v *validator.Validate,
	ss *services.SubscriptionService,
) http.Handler {
	sh := handlers.NewSubscriptionHandler(*ss, *v)
	return routers.HandleRequests(sh)
}

func createConsumers(ss *services.SubscriptionService) *consumers.SubscriptionConsumer {
	sc := consumers.NewConsumer(ss)
	return sc
}

func generateCreds() (credentials.TransportCredentials, error) {
	cleanPath := filepath.Clean(rootCACertFilePath)

	if !strings.HasPrefix(cleanPath, "/certs/") {
		return nil, errors.New("invalid certificate path")
	}

	pemData, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read root cert file at %s: %w", rootCACertFilePath, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemData) {
		return nil, errors.New("failed to add CA to pool")
	}

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: "content-service",
		MinVersion: tls.VersionTLS13,
	}
	return credentials.NewTLS(tlsConfig), nil
}
