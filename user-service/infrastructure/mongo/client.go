package mongo

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const timeout = 10 * time.Second

// InitMongoClient establishes a MongoDB client using environment configuration.
func InitMongoClient() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	uri := "mongodb://" + net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMonitor(otelmongo.NewMonitor())
	clientOptions.SetAuth(options.Credential{
		Username:   os.Getenv("DB_USER"),
		Password:   os.Getenv("DB_PASS"),
		AuthSource: os.Getenv("DB_NAME"),
	})

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		// Ignoring disconnect errors to prioritize `ping` failure
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping serve: %w", err)
	}

	return client, nil
}
