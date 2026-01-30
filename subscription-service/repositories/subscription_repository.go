package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/subscriptions/entities"
	"go.mongodb.org/mongo-driver/mongo"
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

func (r *SubscriptionRepository) Create(s *entities.Subscription, ctx context.Context) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, s)
	if err != nil {
		return err
	}

	return nil
}
