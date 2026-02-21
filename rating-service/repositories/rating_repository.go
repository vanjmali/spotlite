package repositories

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/rating-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrRatingAlreadyExists = errors.New("user has already rated the given content")
)

type RatingRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *RatingRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

func NewRatingRepository(dbName string, collName string, c *mongo.Client) *RatingRepository {
	r := RatingRepository{Client: c, DbName: dbName, CollName: collName}

	return &r
}

func (r *RatingRepository) Create(rating *entities.Rating, ctx context.Context) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, rating)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return ErrRatingAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (r *RatingRepository) Delete(entityID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error) {
	c := r.getCollection()

	res, err := c.DeleteOne(ctx, bson.M{
		"user_id":   userID,
		"entity_id": entityID,
	})
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}
