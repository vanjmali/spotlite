package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/vanjmali/spotlite/rating-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrRatingAlreadyExists = errors.New("user has already rated the given content")
	ErrFindRatings         = errors.New("error has occured while finding ratings for the given parameters")
	ErrRatingCursor        = errors.New("error has occured while loading rating cursor")
	ErrUUIDParse           = errors.New("error has occurred while parsing song IDs")
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

func (r *RatingRepository) Delete(ratingID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error) {
	c := r.getCollection()

	res, err := c.DeleteOne(ctx, bson.M{
		"_id":     ratingID,
		"user_id": userID,
	})
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

func (r *RatingRepository) FindRatingsBySongID(
	ctx context.Context,
	songID primitive.ObjectID,
	batchSize int,
	lastID *primitive.ObjectID,
) ([]*entities.Rating, string, error) {
	c := r.getCollection()

	filter := bson.M{"song_id": songID}

	if lastID != nil {
		filter["_id"] = bson.M{"$lt": *lastID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(batchSize))

	cursor, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrFindRatings, err)
	}
	defer cursor.Close(ctx)

	ratings := make([]*entities.Rating, 0)
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrRatingCursor, err)
	}

	var nextID string
	if len(ratings) > 0 {
		nextID = ratings[len(ratings)-1].ID.Hex()
	}

	return ratings, nextID, nil
}

func (r *RatingRepository) FindRatingsByUserID(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Rating, int64, error) {
	c := r.getCollection()

	total, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}})

	cur, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	ratings := make([]entities.Rating, 0)
	if err = cur.All(ctx, &ratings); err != nil {
		return nil, 0, err
	}
	return ratings, total, nil
}
