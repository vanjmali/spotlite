package repositories

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/subscription-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrSubscriptionAlreadyExists = errors.New("user is already subscribed to the given content")

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
