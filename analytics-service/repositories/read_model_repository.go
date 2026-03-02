package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/vanjmali/spotlite/analytics-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrReadModelNotFound = errors.New("read model not found")
)

// UserAnalyticsRepository provides data access helpers for CQRS read models.
// It persists denormalized analytics aggregates for efficient queries.
type UserAnalyticsRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *UserAnalyticsRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewUserAnalyticsRepository constructs a UserAnalyticsRepository for the given database and collection.
func NewUserAnalyticsRepository(dbName string, collName string, c *mongo.Client) *UserAnalyticsRepository {
	r := UserAnalyticsRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// EnsureIndexes creates indexes required for efficient read model queries.
// Indexes support the following query patterns:
// - Find analytics by user for read model queries
func (r *UserAnalyticsRepository) EnsureIndexes(ctx context.Context) error {
	c := r.getCollection()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_analytics_user_id_idx").SetUnique(true),
		},
	}

	_, err := c.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create user analytics indexes: %w", err)
	}

	return nil
}

// UpsertUserAnalytics creates or updates a user analytics aggregate.
// Used by read model projections to maintain denormalized analytics data.
func (r *UserAnalyticsRepository) UpsertUserAnalytics(
	ctx context.Context,
	analytics *entities.UserAnalyticsReadModel,
) error {
	if analytics == nil {
		return ErrInvalidEvent
	}

	c := r.getCollection()

	filter := bson.M{"user_id": analytics.UserID}
	update := bson.M{
		"$set": bson.M{
			"total_songs_played":       analytics.TotalSongsPlayed,
			"average_rating":           analytics.AverageRating,
			"rating_sum":               analytics.RatingSum,
			"ratings_count":            analytics.RatingsCount,
			"songs_by_genre":           analytics.SongsByGenre,
			"top_artists":              analytics.TopArtists,
			"subscribed_artists_count": analytics.SubscribedArtistsCount,
		},
		"$setOnInsert": bson.M{
			"user_id": analytics.UserID,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert user analytics: %w", err)
	}

	return nil
}

// GetUserAnalytics retrieves analytics aggregates for a specific user.
// Returns ErrReadModelNotFound if the user has no analytics record.
func (r *UserAnalyticsRepository) GetUserAnalytics(
	ctx context.Context,
	userID string,
) (*entities.UserAnalyticsReadModel, error) {
	c := r.getCollection()

	filter := bson.M{"user_id": userID}

	var analytics entities.UserAnalyticsReadModel
	err := c.FindOne(ctx, filter).Decode(&analytics)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrReadModelNotFound
		}
		return nil, fmt.Errorf("failed to query user analytics: %w", err)
	}

	return &analytics, nil
}
