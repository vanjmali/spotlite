package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/recommendation-service/consumers"
	"github.com/vanjmali/spotlite/recommendation-service/handlers"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
	"github.com/vanjmali/spotlite/recommendation-service/routers"
	"github.com/vanjmali/spotlite/recommendation-service/services"
)

const (
	ratingCreatedDurable       = events.RATING_DURABLE + "_CREATED"
	ratingUpdatedDurable       = events.RATING_DURABLE + "_UPDATED"
	ratingDeletedDurable       = events.RATING_DURABLE + "_DELETED"
	subscriptionCreatedDurable = events.SUBSCRIPTION_DURABLE + "_CREATED"
	subscriptionDeletedDurable = events.SUBSCRIPTION_DURABLE + "_DELETED"
)

var (
	certFilePath = utils.MustGetEnv("CERT_PATH")
	keyFilePath  = utils.MustGetEnv("KEY_PATH")
	config       = server.ServerRunConfiguration{
		TelemetryName: "recommendation-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			dbc, jsc, err := createClients()
			if err != nil {
				err = fmt.Errorf("failed to create clients: %w", err)
				return h, shutdown, err
			}

			// Cleanup resources on error
			defer func() {
				if err == nil {
					return
				}
				jsc.Close()
				_ = dbc.Close(ctx)
			}()

			// Ensure streams exist
			if err = jsc.EnsureStream(ctx, events.RATINGS_STREAM, []string{
				events.SUBJECT_RATING_CREATED,
				events.SUBJECT_RATING_UPDATED,
				events.SUBJECT_RATING_DELETED,
			}); err != nil {
				err = fmt.Errorf("failed to ensure ratings stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(ctx, events.LISTENS_STREAM, []string{
				events.SUBJECT_LISTEN_CREATED,
			}); err != nil {
				err = fmt.Errorf("failed to ensure listens stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(ctx, events.SUBSCRIPTIONS_STREAM, []string{
				events.SUBJECT_SUBSCRIPTION_CREATED,
				events.SUBJECT_SUBSCRIPTION_DELETED,
			}); err != nil {
				err = fmt.Errorf("failed to ensure subscriptions stream: %w", err)
				return h, shutdown, err
			}

			ur, sr, ar, gr, abr, rr := createRepositories(dbc)
			ss := createServices(ur, sr, ar, gr, abr, rr)
			h = createHandlers(ss)
			c := createConsumers(ur, sr, ar, gr, abr, rr)

			// Start consumers in background.
			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup
			consumerErrCh := make(chan error, 6)

			startConsumer := func(stream, subject, durable, label string, handler events.SubscribeHandler) {
				consumerWg.Add(1)
				go func() {
					defer consumerWg.Done()
					err := jsc.StartConsumer(consumerCtx, stream, subject, durable, handler)
					if err != nil && !errors.Is(err, context.Canceled) {
						logging.Errorf(ctx, "%s consumer error: %v", label, err)
						consumerErrCh <- fmt.Errorf("%s consumer error: %w", label, err)
					}
				}()
			}

			startConsumer(
				events.RATINGS_STREAM,
				events.SUBJECT_RATING_CREATED,
				ratingCreatedDurable,
				"rating created",
				c.HandleRatingCreated,
			)
			startConsumer(
				events.RATINGS_STREAM,
				events.SUBJECT_RATING_UPDATED,
				ratingUpdatedDurable,
				"rating updated",
				c.HandleRatingUpdated,
			)
			startConsumer(
				events.RATINGS_STREAM,
				events.SUBJECT_RATING_DELETED,
				ratingDeletedDurable,
				"rating deleted",
				c.HandleRatingDeleted,
			)
			startConsumer(
				events.LISTENS_STREAM,
				events.SUBJECT_LISTEN_CREATED,
				events.LISTEN_DURABLE,
				"listen created",
				c.HandleListenCreated,
			)
			startConsumer(
				events.SUBSCRIPTIONS_STREAM,
				events.SUBJECT_SUBSCRIPTION_CREATED,
				subscriptionCreatedDurable,
				"subscription created",
				c.HandleSubscriptionCreated,
			)
			startConsumer(
				events.SUBSCRIPTIONS_STREAM,
				events.SUBJECT_SUBSCRIPTION_DELETED,
				subscriptionDeletedDurable,
				"subscription deleted",
				c.HandleSubscriptionDeleted,
			)

			shutdown = func() error {
				var errs []error

				consumerCancel()
				consumerWg.Wait()
				close(consumerErrCh)
				for consumerErr := range consumerErrCh {
					errs = append(errs, consumerErr)
				}

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				jsc.Close()
				if err := dbc.Close(shutdownCtx); err != nil {
					errs = append(errs, fmt.Errorf("failed to close Neo4j driver: %w", err))
				}

				if len(errs) > 0 {
					return fmt.Errorf("shutdown errors: %v", errs)
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

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		logging.Errorf(context.Background(), "failed to start recommendation service: %v", err)
	}
}

func createClients() (neo4j.DriverWithContext, *events.JetStreamClient, error) {
	neoUri := utils.MustGetEnv("NEO4J_URI")
	neoUser := utils.MustGetEnv("NEO4J_USER")
	neoPass := utils.MustGetEnv("NEO4J_PASSWORD")

	dbc, err := neo4j.NewDriverWithContext(neoUri, neo4j.BasicAuth(neoUser, neoPass, ""))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	natsUrl := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	jsc, err := events.NewClient(natsUrl)
	if err != nil {
		_ = dbc.Close(context.Background())
		return nil, nil, fmt.Errorf("failed to create NATS JetStream client: %w", err)
	}

	return dbc, jsc, nil
}

func createRepositories(driver neo4j.DriverWithContext) (
	*repositories.UserNodeRepository,
	*repositories.SongNodeRepository,
	*repositories.ArtistNodeRepository,
	*repositories.GenreNodeRepository,
	*repositories.AlbumNodeRepository,
	*repositories.GraphRelationRepository,
) {
	ur := repositories.NewUserNodeRepository(driver)
	sr := repositories.NewSongNodeRepository(driver)
	ar := repositories.NewArtistNodeRepository(driver)
	gr := repositories.NewGenreNodeRepository(driver)
	abr := repositories.NewAlbumNodeRepository(driver)
	rr := repositories.NewGraphRelationRepository(driver)

	return ur, sr, ar, gr, abr, rr
}

func createServices(
	ur *repositories.UserNodeRepository,
	sr *repositories.SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
	abr *repositories.AlbumNodeRepository,
	rr *repositories.GraphRelationRepository,
) *services.Services {
	return services.NewServices(ur, sr, ar, gr, abr, rr)
}

func createConsumers(
	ur *repositories.UserNodeRepository,
	sr *repositories.SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
	abr *repositories.AlbumNodeRepository,
	rr *repositories.GraphRelationRepository,
) *consumers.RecommendationConsumer {
	return consumers.NewRecommendationConsumer(ur, sr, ar, gr, abr, rr)
}

func createHandlers(ss *services.Services) http.Handler {
	rh := handlers.NewRecommendationHandler(*ss)
	return routers.HandleRequests(rh)
}
