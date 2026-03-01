package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserActivityRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func NewUserActivityRepository(dbName string, collName string, c *mongo.Client) *UserActivityRepository {
	r := UserActivityRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *UserActivityRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

func (r *UserActivityRepository) EnsureIndexes(ctx context.Context) error {
	c := r.getCollection()
	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetName("activity_event_id_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "occurred_at", Value: -1}},
			Options: options.Index().SetName("activity_user_occurred_at_desc"),
		},
	})
	return err
}

func (r *UserActivityRepository) Append(ctx context.Context, activity entities.UserActivity) error {
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now().UTC()
	}

	_, err := r.getCollection().InsertOne(ctx, activity)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert user activity: %w", err)
	}

	return nil
}

func (r *UserActivityRepository) ListByUserID(
	ctx context.Context,
	userID string,
	skip int64,
	limit int64,
) ([]entities.UserActivity, int64, error) {
	filter := bson.M{"user_id": userID}

	total, err := r.getCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user activity: %w", err)
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "occurred_at", Value: -1}})
	cur, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user activity: %w", err)
	}
	defer cur.Close(ctx)

	items := make([]entities.UserActivity, 0)
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to decode user activity: %w", err)
	}

	return items, total, nil
}
