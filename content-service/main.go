package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	commonvalid "github.com/vanjmali/spotlite/common-lib/validations"
	"github.com/vanjmali/spotlite/content/handlers"
	"github.com/vanjmali/spotlite/content/infrastructure/mongo"
	"github.com/vanjmali/spotlite/content/repositories"
	"github.com/vanjmali/spotlite/content/routers"
	"github.com/vanjmali/spotlite/content/services"
	contentvalid "github.com/vanjmali/spotlite/content/validations"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "content-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
	ConfigureValidation: func(v *validator.Validate) error {
		requests.RegisterJSONTagNameFunc(v)
		if err := requests.RegisterValidation(v, commonvalid.CheckValidName); err != nil {
			return fmt.Errorf("failed to register name validation: %w", err)
		}
		if err := requests.RegisterValidation(v, contentvalid.CheckValidDateOnly); err != nil {
			return fmt.Errorf("failed to register date-only validation: %w", err)
		}

		return nil
	},
	CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
		dbc, err := createClients()
		if err != nil {
			err = fmt.Errorf("failed to create clients: %w", err)
			return h, shutdown, err
		}

		// Cleanup resources on error
		defer func() {
			if err == nil {
				return
			}
			_ = dbc.Disconnect(ctx)
		}()

		ar, sr, alr, gr := createRepositories(dbc)
		gs, as, ss, als := createServices(ar, sr, alr, gr)
		h = createHandlers(v, as, ss, als, gs)

		shutdown = func() error {
			if err := dbc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}
			return nil
		}

		return h, shutdown, err
	},
}

func createClients() (*mongodriver.Client, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	return dbc, nil
}

func createRepositories(dbc *mongodriver.Client) (
	*repositories.ArtistRepository,
	*repositories.SongRepository,
	*repositories.AlbumRepository,
	*repositories.GenreRepository,
) {
	name := utils.MustGetEnv("DB_NAME")
	ar := repositories.NewArtistRepository(name, "artists", dbc)
	sr := repositories.NewSongRepository(name, "songs", dbc)
	alr := repositories.NewAlbumRepository(name, "albums", dbc)
	gr := repositories.NewGenreRepository(name, "genres", dbc)

	return ar, sr, alr, gr
}

func createServices(
	ar *repositories.ArtistRepository,
	sr *repositories.SongRepository,
	alr *repositories.AlbumRepository,
	gr *repositories.GenreRepository,
) (
	*services.GenreService,
	*services.ArtistService,
	*services.SongService,
	*services.AlbumService,
) {
	gs := services.NewGenreService(*gr)
	as := services.NewArtistService(*ar, *gs)
	ss := services.NewSongService(*sr, *as, *gs)
	als := services.NewAlbumService(*alr, *as, *ss, *gs)

	return gs, as, ss, als
}

func createHandlers(
	v *validator.Validate,
	as *services.ArtistService,
	ss *services.SongService,
	als *services.AlbumService,
	gs *services.GenreService,
) http.Handler {
	ah := handlers.NewArtistHandler(*as, *v)
	sh := handlers.NewSongHandler(*ss, *v)
	alh := handlers.NewAlbumHandler(*als, *v)
	gh := handlers.NewGenreHandler(*gs, *v)
	gsh := handlers.NewGlobalSearchHandler(*gs, *ss, *als, *as)

	return routers.HandleRequests(ah, sh, alh, gh, gsh)
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start content service: %v", err)
	}
}
