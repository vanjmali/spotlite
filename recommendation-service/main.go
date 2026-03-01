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
	certFilePath       = utils.MustGetEnv("CERT_PATH")
	keyFilePath        = utils.MustGetEnv("KEY_PATH")
	rootCACertFilePath = utils.MustGetEnv("ROOT_CERT_PATH")
	natsURL            = utils.MustGetEnv("NATS_URL")
	config             = server.ServerRunConfiguration{
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

			if err = jsc.EnsureStream(ctx, events.USERS_STREAM, []string{events.SUBJECT_USER_CREATED}); err != nil {
				err = fmt.Errorf("failed to ensure users stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(
				ctx,
				events.SONGS_STREAM,
				[]string{events.SUBJECT_SONG_CREATED, events.SUBJECT_SONG_RATED, events.SUBJECT_SONG_UPDATED, events.SUBJECT_SONG_DELETED},
			); err != nil {
				err = fmt.Errorf("failed to ensure songs stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(
				ctx,
				events.GENRES_STREAM,
				[]string{events.SUBJECT_GENRE_SUBSCRIBED, events.SUBJECT_GENRE_CREATED, events.SUBJECT_GENRE_UPDATED},
			); err != nil {
				err = fmt.Errorf("failed to ensure genres stream: %w", err)
				return h, shutdown, err
			}

			ur, gr, rr := createRepositories(dbc)
			_, rs := createServices(ur, gr, rr)
			h = createHandlers(rs)
			c := createConsumers(rs)

			// Start consumers in background.
			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup
			consumerErrCh := make(chan error, 8)

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
				events.GENRES_STREAM,
				events.SUBJECT_GENRE_CREATED,
				events.GENRE_CREATE_DURABLE,
				"genre created",
				c.HandleGenreCreation,
			)

			startConsumer(
				events.GENRES_STREAM,
				events.SUBJECT_GENRE_SUBSCRIBED,
				events.GENRE_SUB_DURABLE,
				"genre subscription created",
				c.HandleGenreSubscription,
			)

			startConsumer(
				events.GENRES_STREAM,
				events.SUBJECT_GENRE_UPDATED,
				events.GENRE_UPDATE_DURABLE,
				"genre updated",
				c.HandleGenreUpdate,
			)

			startConsumer(
				events.SONGS_STREAM,
				events.SUBJECT_SONG_UPDATED,
				events.SONG_UPDATE_DURABLE,
				"song updated",
				c.HandleSongUpdate,
			)

			startConsumer(
				events.SONGS_STREAM,
				events.SUBJECT_SONG_DELETED,
				events.SONG_DELETE_DURABLE_RECOMMENDATION,
				"song deleted",
				c.HandleSongDelete,
			)

			startConsumer(
				events.SONGS_STREAM,
				events.SUBJECT_SONG_CREATED,
				events.SONG_CREATE_DURABLE,
				"song created",
				c.HandleSongCreation,
			)

			startConsumer(
				events.SONGS_STREAM,
				events.SUBJECT_SONG_RATED,
				events.SONG_RATE_DURABLE,
				"song rating created",
				c.HandleSongRating,
			)

			startConsumer(
				events.USERS_STREAM,
				events.SUBJECT_USER_CREATED,
				events.USER_DURABLE,
				"user created",
				c.HandleUserRegistration,
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
		logging.Errorf(context.Background(), "failed to start recommendation service: %v", err)
	}
}

func ensureConstraints(ctx context.Context, d neo4j.DriverWithContext) error {
	session := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// a list of constraints to apply
	queries := []string{
		"CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (u:User) REQUIRE u.user_id IS UNIQUE",
		"CREATE CONSTRAINT genre_id_unique IF NOT EXISTS FOR (g:Genre) REQUIRE g.genre_id IS UNIQUE",
		"CREATE CONSTRAINT song_id_unique IF NOT EXISTS FOR (s:Song) REQUIRE s.song_id IS UNIQUE",
	}

	for _, query := range queries {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}
			return result.Consume(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to apply constraint [%s]: %w", query, err)
		}
	}

	return nil
}

func createClients() (neo4j.DriverWithContext, *events.JetStreamClient, error) {
	neoUri := utils.MustGetEnv("NEO4J_URI")
	neoUser := utils.MustGetEnv("NEO4J_USER")
	neoPass := utils.MustGetEnv("NEO4J_PASSWORD")

	dbc, err := neo4j.NewDriverWithContext(neoUri, neo4j.BasicAuth(neoUser, neoPass, ""))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	err = ensureConstraints(context.Background(), dbc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ensure constraints: %w", err)
	}

	jsc, err := events.NewClient(natsURL, nats.RootCAs(rootCACertFilePath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialized NATS jet stream client: %w", err)
	}

	return dbc, jsc, nil
}

func createRepositories(driver neo4j.DriverWithContext) (
	*repositories.UserNodeRepository,
	*repositories.GenreNodeRepository,
	*repositories.GraphRelationRepository,
) {
	ur := repositories.NewUserNodeRepository(driver)
	gr := repositories.NewGenreNodeRepository(driver)
	rr := repositories.NewGraphRelationRepository(driver)

	return ur, gr, rr
}

func createServices(
	ur *repositories.UserNodeRepository,
	gr *repositories.GenreNodeRepository,
	rr *repositories.GraphRelationRepository,
) (*services.Repositories, *services.RecommendationService) {
	baseServices := services.NewServices(ur, gr, rr)
	recommendationService := services.NewRecommendationService(baseServices)
	return baseServices, recommendationService
}

func createConsumers(
	rs *services.RecommendationService,
) *consumers.RecommendationConsumer {
	return consumers.NewRecommendationConsumer(rs)
}

func createHandlers(rs *services.RecommendationService) http.Handler {
	rh := handlers.NewRecommendationHandler(rs)
	return routers.HandleRequests(rh)
}
