package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
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
	pb "github.com/vanjmali/spotlite/common-lib/proto/rating_service"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/rating-service/consumers"
	"github.com/vanjmali/spotlite/rating-service/handlers"
	adapters "github.com/vanjmali/spotlite/rating-service/infrastructure/grpc"
	"github.com/vanjmali/spotlite/rating-service/infrastructure/mongo"
	"github.com/vanjmali/spotlite/rating-service/repositories"
	"github.com/vanjmali/spotlite/rating-service/routers"
	"github.com/vanjmali/spotlite/rating-service/services"
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
	natsURL            = utils.MustGetEnv("NATS_URL")
	config             = server.ServerRunConfiguration{
		TelemetryName: "rating-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		ConfigureValidation: func(v *validator.Validate) error {
			return nil
		},
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			dbc, gc, jsc, err := createClients()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create clients: %w", err)
			}

			var grpcServer *grpc.Server

			defer func() {
				if err == nil {
					return
				}
				if dbc != nil {
					_ = dbc.Disconnect(context.Background())
				}
				if jsc != nil {
					jsc.Close()
				}
				if gc != nil {
					_ = gc.Close()
				}
				if grpcServer != nil {
					grpcServer.Stop()
				}
			}()

			if err = initializeRatingIndexes(ctx, dbc); err != nil {
				return nil, nil, err
			}

			gcc := createAdapters(gc)
			rr := createRepositories(dbc)
			rs := createServices(rr, gcc, jsc)
			c := createConsumers(rs)

			grpcServer, err = createGrpcServer(rs)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create grpc server: %w", err)
			}

			grpcPort := utils.GetEnv("GRPC_PORT", "50051")
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to listen on grpc port: %w", err)
			}

			go func() {
				log.Printf("gRPC rating server listening on port %s", grpcPort)
				if serveErr := grpcServer.Serve(lis); serveErr != nil {
					log.Printf("failed to serve rating grpc: %v", serveErr)
				}
			}()
			if err = jsc.EnsureStream(ctx, events.RATINGS_STREAM, []string{
				events.SUBJECT_RATING_CREATED,
				events.SUBJECT_RATING_UPDATED},
			); err != nil {
				return nil, nil, fmt.Errorf("failed to ensure ratings stream: %w", err)
			}

			if err = jsc.EnsureStream(ctx, events.SONGS_STREAM, []string{
				events.SUBJECT_SONG_CREATED,
				events.SUBJECT_SONG_UPDATED,
				events.SUBJECT_SONG_DELETED,
			}); err != nil {
				return nil, nil, fmt.Errorf("failed to ensure songs stream: %w", err)
			}

			consumerCtx, consumerCancel := context.WithCancel(ctx)
			var consumerWg sync.WaitGroup

			consumerErrCh := make(chan error, 5)

			startConsumer := func(stream, subject, durable, label string, handler events.SubscribeHandler) {
				consumerWg.Add(1)
				go func() {
					defer consumerWg.Done()
					err := jsc.StartConsumer(consumerCtx, stream, subject, durable, handler)
					if err != nil && !errors.Is(err, context.Canceled) {
						logging.Errorf(ctx, "%s consumer error: %v", label, err)

						select {
						case consumerErrCh <- fmt.Errorf("%s consumer error: %w", label, err):
						default:
						}
					}
				}()
			}

			startConsumer(
				events.SONGS_STREAM,
				events.SUBJECT_SONG_DELETED,
				events.SONG_DELETE_DURABLE_RATING,
				"song deleted",
				c.HandleSongDelete,
			)

			h = createHandlers(v, rs)

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

				if gc != nil {
					if err := gc.Close(); err != nil {
						errs = append(errs, fmt.Errorf("grpc client close error: %w", err))
					}
				}

				if grpcServer != nil {
					grpcServer.GracefulStop()
				}

				if jsc != nil {
					jsc.Close()
				}

				if dbc != nil {
					if err := dbc.Disconnect(shutdownCtx); err != nil {
						errs = append(errs, fmt.Errorf("failed to disconnect mongo client: %w", err))
					}
				}

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

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start rating service: %v", err)
	}
}

func createClients() (*mongodriver.Client, *grpc.ClientConn, *events.JetStreamClient, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
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

	jsc, err := events.NewClient(natsURL, nats.RootCAs(rootCACertFilePath))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize NATS JetStream client: %w", err)
	}

	return dbc, gc, jsc, nil
}

func initializeRatingIndexes(ctx context.Context, c *mongodriver.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models := []mongodriver.IndexModel{
		{
			Keys:    bson.D{{Key: "song_id", Value: 1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("idx_song_id_newest"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("idx_user_id_newest"),
		},
		{
			Keys: bson.D{
				{Key: "song_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_song_user"),
		},
	}

	_, err := c.Database(mongo.DatabaseName()).
		Collection("ratings").
		Indexes().
		CreateMany(ctx, models)

	return err
}

func createAdapters(gc *grpc.ClientConn) *adapters.GrpcContentEntityGetter {
	return adapters.NewGrpcContentEntityGetter(gc)
}

func createRepositories(dbc *mongodriver.Client) *repositories.RatingRepository {
	name := utils.MustGetEnv("DB_NAME")
	sr := repositories.NewRatingRepository(name, "ratings", dbc)

	return sr
}

func createServices(
	sr *repositories.RatingRepository,
	gcc *adapters.GrpcContentEntityGetter,
	jsc *events.JetStreamClient,
) *services.RatingService {
	rs := services.NewRatingService(sr, gcc, jsc)

	return rs
}

func createHandlers(
	v *validator.Validate,
	rs *services.RatingService,
) http.Handler {
	rh := handlers.NewRatingHandler(*rs, *v)
	return routers.HandleRequests(rh)
}

func createGrpcServer(rs *services.RatingService) (*grpc.Server, error) {
	ratingGrpcServer := adapters.NewRatingServer(rs)

	creds, err := credentials.NewServerTLSFromFile(certFilePath, keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS keys: %w", err)
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterGetSongRatingServer(s, ratingGrpcServer)

	return s, nil
}

func createConsumers(
	rs *services.RatingService,
) *consumers.RatingConsumer {
	return consumers.NewRatingConsumer(rs)
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
