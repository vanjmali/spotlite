package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: remove hardcoded url
const connectionString = "mongodb://localhost:27017/"

func InitMongoClient() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connectionString)
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
