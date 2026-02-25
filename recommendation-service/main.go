package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
			ss, rs := createServices(ur, sr, ar, gr, abr, rr)
			h = createHandlers(ss, rs)
			c := createConsumers(ur, sr, ar, gr, abr, rr)

			// Start consumers in background
			consumerCtx, consumerCancel := context.WithCancel(ctx)
			consumerDone := make(chan struct{})

			var consumerErr error

			go func() {
				consumerErr = jsc.StartConsumer(
					consumerCtx,
					events.RATINGS_STREAM,
					events.SUBJECT_RATING_CREATED,
					events.RATING_DURABLE,
					c.HandleRatingCreated,
				)
				if consumerErr != nil {
					logging.Errorf(ctx, "rating created consumer error: %v", consumerErr)
				}
			}()

			go func() {
				err := jsc.StartConsumer(
					consumerCtx,
					events.RATINGS_STREAM,
					events.SUBJECT_RATING_UPDATED,
					events.RATING_DURABLE,
					c.HandleRatingUpdated,
				)
				if err != nil {
					logging.Errorf(ctx, "rating updated consumer error: %v", err)
				}
			}()

			go func() {
				err := jsc.StartConsumer(
					consumerCtx,
					events.RATINGS_STREAM,
					events.SUBJECT_RATING_DELETED,
					events.RATING_DURABLE,
					c.HandleRatingDeleted,
				)
				if err != nil {
					logging.Errorf(ctx, "rating deleted consumer error: %v", err)
				}
			}()

			go func() {
				err := jsc.StartConsumer(
					consumerCtx,
					events.LISTENS_STREAM,
					events.SUBJECT_LISTEN_CREATED,
					events.LISTEN_DURABLE,
					c.HandleListenCreated,
				)
				if err != nil {
					logging.Errorf(ctx, "listen consumer error: %v", err)
				}
			}()

			go func() {
				err := jsc.StartConsumer(
					consumerCtx,
					events.SUBSCRIPTIONS_STREAM,
					events.SUBJECT_SUBSCRIPTION_CREATED,
					events.SUBSCRIPTION_DURABLE,
					c.HandleSubscriptionCreated,
				)
				if err != nil {
					logging.Errorf(ctx, "subscription created consumer error: %v", err)
				}
			}()

			go func() {
				err := jsc.StartConsumer(
					consumerCtx,
					events.SUBSCRIPTIONS_STREAM,
					events.SUBJECT_SUBSCRIPTION_DELETED,
					events.SUBSCRIPTION_DURABLE,
					c.HandleSubscriptionDeleted,
				)
				if err != nil {
					logging.Errorf(ctx, "subscription deleted consumer error: %v", err)
				}
				close(consumerDone)
			}()

			go func() {
				<-consumerDone
				if consumerErr != nil && !errors.Is(consumerErr, context.Canceled) {
					logging.Errorf(context.Background(), "recommendation consumer stopped unexpectedly: %v", consumerErr)
				}
			}()

			shutdown = func() error {
				var errs []error

				consumerCancel()
				<-consumerDone

				if consumerErr != nil && !errors.Is(consumerErr, context.Canceled) {
					errs = append(errs, fmt.Errorf("recommendation consumer error: %w", consumerErr))
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
	neoUri := utils.GetEnv("SRV_REC_NEO4J_URI", "neo4j://localhost:7687")
	neoUser := utils.GetEnv("SRV_REC_NEO4J_USER", "neo4j")
	neoPass := utils.GetEnv("SRV_REC_NEO4J_PASSWORD", "password")

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
) (*services.Services, *services.RecommendationService) {
	baseServices := services.NewServices(ur, sr, ar, gr, abr, rr)
	recommendationService := services.NewRecommendationService(baseServices)
	return baseServices, recommendationService
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

func createHandlers(ss *services.Services, rs *services.RecommendationService) http.Handler {
	rh := handlers.NewRecommendationHandler(rs)
	return routers.HandleRequests(rh)
}
