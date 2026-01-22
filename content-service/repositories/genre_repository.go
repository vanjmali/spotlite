package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GenreRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *GenreRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewRepository constructs a GenreRepository for the given database and collection.
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
