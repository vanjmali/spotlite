package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GenreRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *GenreRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewGenreRepository constructs a GenreRepository for the given database and collection.
func NewGenreRepository(dbName string, collName string, c *mongo.Client) *GenreRepository {
	r := GenreRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *GenreRepository) Create(ctx context.Context, genre entities.Genre) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, genre)
	if err != nil {
		return err
	}

	return nil
}

func (r *GenreRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Genre, error) {
	c := r.getCollection()

	var genre entities.Genre

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&genre); err != nil {
		return nil, err
	}
	return &genre, nil
}

func (r *GenreRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update map[string]any) (*entities.Genre, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedGenre entities.Genre
	if err := c.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedGenre); err != nil {
		return nil, err
	}

	return &updatedGenre, nil
}

func (r *GenreRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	res, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *GenreRepository) FindAll(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Genre, int64, error) {
	c := r.getCollection()

	total, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cur, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	genres := make([]entities.Genre, 0)
	if err := cur.All(ctx, &genres); err != nil {
		return nil, 0, err
	}

	return genres, total, nil
}
