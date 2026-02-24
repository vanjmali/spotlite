package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
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
			dbc, err := createClients()
			if err != nil {
				err = fmt.Errorf("failed to create Neo4j client: %w", err)
				return h, shutdown, err
			}

			// Cleanup resources on error
			defer func() {
				if err == nil {
					return
				}
				_ = dbc.Close(ctx)
			}()

			ur, sr, ar, gr, abr, rr := createRepositories(dbc)
			ss := createServices(ur, sr, ar, gr, abr, rr)
			h = createHandlers(ss)

			shutdown = func() error {
				if err := dbc.Close(ctx); err != nil {
					return fmt.Errorf("failed to close Neo4j driver: %w", err)
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

func createClients() (neo4j.DriverWithContext, error) {
	neoUri := utils.GetEnv("NEO4J_URI", "neo4j://localhost:7687")
	neoUser := utils.GetEnv("NEO4J_USER", "neo4j")
	neoPass := utils.GetEnv("NEO4J_PASSWORD", "password")

	return neo4j.NewDriverWithContext(neoUri, neo4j.BasicAuth(neoUser, neoPass, ""))
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

func createHandlers(ss *services.Services) http.Handler {
	rh := handlers.NewRecommendationHandler(*ss)
	return routers.HandleRequests(rh)
}
