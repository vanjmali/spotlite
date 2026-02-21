package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
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
	config             = server.ServerRunConfiguration{
		TelemetryName: "rating-service",
		Port:          utils.GetEnv("APP_PORT", "3000"),
		ConfigureValidation: func(v *validator.Validate) error {

			return nil
		},
		CreateHandler: func(ctx context.Context, v *validator.Validate) (h http.Handler, shutdown func() error, err error) {
			dbc, gc, err := createClients()
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

			err = initializeRatingIndexes(ctx, dbc)
			if err != nil {
				return nil, nil, err
			}

			gcc := createAdapters(gc)
			rr := createRepositories(dbc)
			rs := createServices(rr, gcc)
			h = createHandlers(v, rs)
			shutdown = func() error {
				var errs []error

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := gc.Close(); err != nil {
					errs = append(errs, fmt.Errorf("grpc close error: %w", err))
				}

				if err := dbc.Disconnect(shutdownCtx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
					errs = append(errs, fmt.Errorf("failed to disconnect mongo client: %w", err))
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
		log.Fatalf("failed to start rating service: %v", err)
	}
}

func createClients() (*mongodriver.Client, *grpc.ClientConn, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	creds, err := generateCreds()
	if err != nil {
		return nil, nil, err
	}

	grpcTarget := utils.MustGetEnv("CONTENT_GRPC_ADDRESS")

	gc, err := grpc.NewClient(
		grpcTarget,
		grpc.WithTransportCredentials(creds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish a RPC connection with the content-service: %w", err)
	}

	return dbc, gc, nil
}

func initializeRatingIndexes(ctx context.Context, c *mongodriver.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models := []mongodriver.IndexModel{
		{
			Keys:    bson.D{{Key: "song_id", Value: 1}},
			Options: options.Index().SetName("idx_song_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
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
) *services.RatingService {
	rs := services.NewRatingService(sr, gcc)

	return rs
}

func createHandlers(
	v *validator.Validate,
	rs *services.RatingService,
) http.Handler {
	rh := handlers.NewRatingHandler(*rs, *v)
	return routers.HandleRequests(rh)
}

func generateCreds() (credentials.TransportCredentials, error) {
	cleanPath := filepath.Clean(rootCACertFilePath)

	if !strings.HasPrefix(cleanPath, "/certs/") {
		return nil, fmt.Errorf("invalid certificate path")
	}

	pemData, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read root cert file at %s: %w", rootCACertFilePath, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("failed to add CA to pool")
	}

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: "content-service",
		MinVersion: tls.VersionTLS13,
	}
	return credentials.NewTLS(tlsConfig), nil

}
