package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/vanjmali/spotlite/subscription-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrSubscriptionAlreadyExists = errors.New("user is already subscribed to the given content")
	ErrFindSubscriptions         = errors.New("error has occured while finding subscriptions for the given parameters")
	ErrSubscriptionCursor        = errors.New("error has occured while loading subscription cursor")
	ErrUUIDParse                 = errors.New("error has occurred while parsing target IDs")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
)

type SubscriptionRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *SubscriptionRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

func NewSubscriptionRepository(dbName string, collName string, c *mongo.Client) *SubscriptionRepository {
	r := SubscriptionRepository{Client: c, DbName: dbName, CollName: collName}

	return &r
}

func (r *SubscriptionRepository) IsSubscribed(subscriberID, entityID primitive.ObjectID, ctx context.Context) error {
	c := r.getCollection()

	filter := bson.M{"subscriber_id": subscriberID, "entity_id": entityID}
	err := c.FindOne(ctx, filter).Err()
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrSubscriptionNotFound
		default:
			return err
		}
	}

	return nil
}

func (r *SubscriptionRepository) Create(s *entities.Subscription, ctx context.Context) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, s)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return ErrSubscriptionAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (r *SubscriptionRepository) Delete(entityID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error) {
	c := r.getCollection()

	res, err := c.DeleteOne(ctx, bson.M{
		"subscriber_id": userID,
		"entity_id":     entityID,
	})
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

func (r *SubscriptionRepository) FindSubscriptionsByEntityID(
	ctx context.Context,
	targetIDStrs []string,
	batchSize int,
	lastID string,
) ([]*entities.Subscription, string, error) {
	c := r.getCollection()

	targetIDs := make([]primitive.ObjectID, 0, len(targetIDStrs))
	for _, t := range targetIDStrs {
		ID, err := primitive.ObjectIDFromHex(t)
		if err != nil {
			return nil, "", ErrUUIDParse
		}
		targetIDs = append(targetIDs, ID)
	}

	filter := bson.M{
		"entity_id": bson.M{
			"$in": targetIDs,
		},
	}

	if lastID != "" {
		objID, err := primitive.ObjectIDFromHex(lastID)
		if err != nil {
			return nil, "", ErrUUIDParse
		}
		filter["_id"] = bson.M{"$gt": objID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(int64(batchSize))

	cursor, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrFindSubscriptions, err)
	}

	defer cursor.Close(ctx)

	var subs []*entities.Subscription
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrSubscriptionCursor, err)
	}

	var nextID string
	if len(subs) > 0 {
		nextID = subs[len(subs)-1].ID.Hex()
	}

	return subs, nextID, nil
}
