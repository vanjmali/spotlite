package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	commonvalid "github.com/vanjmali/spotlite/common-lib/validations"
	"github.com/vanjmali/spotlite/content/handlers"
	infragrpc "github.com/vanjmali/spotlite/content/infrastructure/grpc"
	"github.com/vanjmali/spotlite/content/infrastructure/mongo"
	"github.com/vanjmali/spotlite/content/repositories"
	"github.com/vanjmali/spotlite/content/routers"
	"github.com/vanjmali/spotlite/content/services"
	"github.com/vanjmali/spotlite/content/storage"
	contentvalid "github.com/vanjmali/spotlite/content/validations"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
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

		hdfsStore, err := storage.NewHDFSStorage()
		if err != nil {
			err = fmt.Errorf("failed to init hdfs client: %w", err)
			_ = dbc.Disconnect(ctx)
			return h, shutdown, err
		}

		if err := hdfsStore.EnsureBaseDir(); err != nil {
			err = fmt.Errorf("failed to ensure hdfs base dir: %w", err)
			_ = dbc.Disconnect(ctx)
			return h, shutdown, err
		}

		// Cleanup resources on error
		defer func() {
			if err == nil {
				return
			}
			_ = dbc.Disconnect(ctx)
			_ = hdfsStore.Close()
		}()

		ar, sr, alr, gr := createRepositories(dbc)
		gs, as, ss, als, glss := createServices(ar, sr, alr, gr, hdfsStore)
		h = createHandlers(v, as, ss, als, gs, glss)

		// configures grpc server
		grpcPort := utils.GetEnv("GRPC_PORT", "50051")
		s := createGrpcServer(gs, as)

		// this doesn't start the server it just reserves the port and prepares everything
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			return h, shutdown, fmt.Errorf("failed to listen on grpc port: %w", err)
		}

		// starts the server in a separate go routine to avoid blocking the http server
		go func() {
			log.Printf("gRPC server listening on port %s", grpcPort)
			if err := s.Serve(lis); err != nil {
				log.Fatalf("failed to serve grpc: %v", err)
			}
		}()

		shutdown = func() error {
			// makes sure to gracefully stop the rpc server
			s.GracefulStop()

			if err := hdfsStore.Close(); err != nil {
				return fmt.Errorf("failed to close hdfs client: %w", err)
			}

			if err := dbc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}
			return nil
		}

		return h, shutdown, err
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start content service: %v", err)
	}
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
	glss *services.GlobalSearchService,
	hdfsStore *storage.HDFSStorage,
) (
	*services.GenreService,
	*services.ArtistService,
	*services.SongService,
	*services.AlbumService,
	*services.GlobalSearchService,
) {
	gs := services.NewGenreService(*gr)
	as := services.NewArtistService(*ar, *gs)
	als := services.NewAlbumService(*alr, *sr, *as, *gs)
	ss := services.NewSongService(*sr, *as, *gs, als, *hdfsStore)
	glss := services.NewGlobalSearchService(gs, ss, als, as)

	return gs, as, ss, als, glss
}

func createHandlers(
	v *validator.Validate,
	as *services.ArtistService,
	ss *services.SongService,
	als *services.AlbumService,
	gs *services.GenreService,
	glss *services.GlobalSearchService,
) http.Handler {
	ah := handlers.NewArtistHandler(*as, *v)
	sh := handlers.NewSongHandler(*ss, *v)
	alh := handlers.NewAlbumHandler(*als, *v)
	gh := handlers.NewGenreHandler(*gs, *v)
	gsh := handlers.NewGlobalSearchHandler(glss)

	return routers.HandleRequests(ah, sh, alh, gh, gsh)
}

func createGrpcServer(gs *services.GenreService, as *services.ArtistService) *grpc.Server {
	// define content grpc server
	contentGrpcServer := infragrpc.NewContentServer(gs, as)

	// this instaniates a new grpc server (engine) which knows how to work with
	// HTTP/2, serialization...
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// make every request that comes to the GetContentEntityServer defined in the proto file
	// be forwarded to the contentGrpcServer instance
	pb.RegisterGetContentEntityServer(s, contentGrpcServer)

	return s
}
