package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/server"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/subscriptions/handlers"
	adapters "github.com/vanjmali/spotlite/subscriptions/infrastructure/grpc"
	"github.com/vanjmali/spotlite/subscriptions/infrastructure/mongo"
	"github.com/vanjmali/spotlite/subscriptions/repositories"
	"github.com/vanjmali/spotlite/subscriptions/routers"
	"github.com/vanjmali/spotlite/subscriptions/services"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var config = server.ServerRunConfiguration{
	TelemetryName: "subscription-service",
	Port:          utils.GetEnv("APP_PORT", "3000"),
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
			_ = gc.Close()
			_ = dbc.Disconnect(ctx)
		}()

		gcc := createAdapters(gc)
		sr := createRepositories(dbc)
		ss := createServices(sr, gcc)
		h = createHandlers(v, ss)

		shutdown = func() error {
			if err := dbc.Disconnect(ctx); err != nil && !errors.Is(err, mongodriver.ErrClientDisconnected) {
				return fmt.Errorf("failed to disconnect mongo client: %w", err)
			}

			if err := gc.Close(); err != nil {
				return fmt.Errorf("failed to close grpc connection: %w", err)
			}
			return nil
		}

		return h, shutdown, err
	},
}

func main() {
	if err := server.Run(context.Background(), config); err != nil {
		log.Fatalf("failed to start subscription service: %v", err)
	}
}

func createClients() (*mongodriver.Client, *grpc.ClientConn, error) {
	dbc, err := mongo.InitMongoClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize subscription service MongoDB client: %w", err)
	}

	grpcTarget := fmt.Sprintf("content-service:%s", utils.GetEnv("CONTENT_GRPC_PORT", "50051"))

	gc, err := grpc.NewClient(
		grpcTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {

		log.Fatalf("did not connect: %v", err)
	}
	return dbc, gc, nil
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
) *services.SubscriptionService {
	ss := services.NewSubscriptionService(sr, gcc)

	return ss
}

func createHandlers(
	v *validator.Validate,
	ss *services.SubscriptionService,
) http.Handler {
	sh := handlers.NewSubscriptionHandler(*ss, *v)

	return routers.HandleRequests(sh)
}
