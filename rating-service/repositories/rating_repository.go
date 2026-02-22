package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/vanjmali/spotlite/rating-service/dtos"
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
)

// RatingRepository provides data access helpers for rating documents.
type RatingRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *RatingRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewRatingRepository constructs a RatingRepository for the given database and collection.
func NewRatingRepository(dbName string, collName string, c *mongo.Client) *RatingRepository {
	r := RatingRepository{Client: c, DbName: dbName, CollName: collName}

	return &r
}

// create func, inserts a new rating into the database.
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

// Delete func, deletes a rating by ID and user ID.
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

// FindByID func, finds a rating by ID.
func (r *RatingRepository) FindByID(ctx context.Context, ratingID primitive.ObjectID) (*entities.Rating, error) {
	c := r.getCollection()

	var rating entities.Rating
	if err := c.FindOne(ctx, bson.M{"_id": ratingID}).Decode(&rating); err != nil {
		return nil, err
	}

	return &rating, nil
}

// FindRatingsBySongID func, finds ratings by song ID with pagination support.
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

// FindRatingsByUserID func, finds ratings by user ID with pagination support.
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

// UpdateByID func, updates a rating by ID and user ID.
func (r *RatingRepository) UpdateByID(
	ctx context.Context,
	ratingID primitive.ObjectID,
	userID primitive.ObjectID,
	update map[string]any,
) (*entities.Rating, error) {
	c := r.getCollection()

	filter := bson.M{"_id": ratingID, "user_id": userID}
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedRating entities.Rating

	if err := c.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedRating); err != nil {
		return nil, err
	}

	return &updatedRating, nil
}

// GetAverageRatingBySongID func, calculates the average rating and count of ratings for a given song ID.
func (r *RatingRepository) GetAverageRatingBySongID(ctx context.Context, songID primitive.ObjectID) (*dtos.SongRatingSummary, error) {
	c := r.getCollection()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"song_id": songID}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$song_id",
			"avg":   bson.M{"$avg": "$value"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cur, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var rows []dtos.SongRatingSummary
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &dtos.SongRatingSummary{Avg: 0, Count: 0}, nil
	}

	return &rows[0], nil
}
