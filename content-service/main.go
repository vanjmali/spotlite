package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
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
	"google.golang.org/grpc/credentials"
)

var (
	rootCACertFilePath = utils.MustGetEnv("ROOT_CERT_PATH")
	certFilePath       = utils.MustGetEnv("CERT_PATH")
	keyFilePath        = utils.MustGetEnv("KEY_PATH")
	config             = server.ServerRunConfiguration{
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
			dbc, jsc, err := createClients()
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

			// make sure stream is already initialized
			err = jsc.EnsureStream(ctx, events.CONTENT_STREAM, []string{events.SUBJECT_ENTITY_CREATED, events.SUBJECT_ENTITY_UPDATED})
			if err != nil {
				err = fmt.Errorf("failed to ensure content stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(ctx, events.SONGS_STREAM, []string{events.SUBJECT_SONG_CREATED, events.SUBJECT_SONG_RATED}); err != nil {
				err = fmt.Errorf("failed to ensure songs stream: %w", err)
				return h, shutdown, err
			}

			if err = jsc.EnsureStream(ctx, events.GENRES_STREAM, []string{events.SUBJECT_GENRE_SUBSCRIBED, events.SUBJECT_GENRE_CREATED}); err != nil {
				err = fmt.Errorf("failed to ensure genres stream: %w", err)
				return h, shutdown, err
			}

			ar, sr, alr, gr := createRepositories(dbc)
			gs, as, ss, als, glss := createServices(ar, sr, alr, gr, jsc, hdfsStore)
			h = createHandlers(v, as, ss, als, gs, glss)

			// configures grpc server
			grpcPort := utils.GetEnv("GRPC_PORT", "50051")
			s, err := createGrpcServer(gs, as, ss, certFilePath, keyFilePath)
			if err != nil {
				return h, shutdown, fmt.Errorf("failed to create grpc server: %w", err)
			}

			// this doesn't start the server it just reserves the port and prepares everything
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				return h, shutdown, fmt.Errorf("failed to listen on grpc port: %w", err)
			}

			// starts the server in a separate go routine to avoid blocking the http server
			go func() {
				logging.Infof(context.Background(), "gRPC server listening on port %s", grpcPort)
				if err := s.Serve(lis); err != nil {
					logging.Errorf(context.Background(), "failed to serve grpc: %v", err)
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

				jsc.Close()

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
			WriteTimeout: 25 * time.Second,
		},
	}
)

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		logging.Errorf(context.Background(), "failed to start content service: %v", err)
	}
}

func createClients() (*mongodriver.Client, *events.JetStreamClient, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	jsc, err := events.NewClient("tls://nats:4222", nats.RootCAs(rootCACertFilePath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialized NATS jets teram client: %w", err)
	}

	return dbc, jsc, nil
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
	jsc *events.JetStreamClient,
	hdfsStore *storage.HDFSStorage,
) (
	*services.GenreService,
	*services.ArtistService,
	*services.SongService,
	*services.AlbumService,
	*services.GlobalSearchService,
) {
	gs := services.NewGenreService(*gr, *jsc)
	as := services.NewArtistService(*ar, *gs, *jsc)
	als := services.NewAlbumService(*alr, *as, *sr, *gs, *jsc)
	ss := services.NewSongService(*sr, *as, *gs, als, hdfsStore)
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

func createGrpcServer(gs *services.GenreService, as *services.ArtistService, ss *services.SongService, cfp, kfp string) (*grpc.Server, error) {
	// define content grpc server
	contentGrpcServer := infragrpc.NewContentServer(gs, as, ss)

	creds, err := credentials.NewServerTLSFromFile(cfp, kfp)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS keys: %w", err)
	}
	// this instaniates a new grpc server (engine) which knows how to work with
	// HTTP/2, serialization...
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// make every request that comes to the GetContentEntityServer defined in the proto file
	// be forwarded to the contentGrpcServer instance
	pb.RegisterGetContentEntityServer(s, contentGrpcServer)

	return s, nil
}
