package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ArtistRepository provides data access helpers for artist documents.
type ArtistRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

// NewRepository constructs a ArtistRepository for the given database and collection.
func NewRepository(dbName string, collName string, c *mongo.Client) *ArtistRepository {
	r := ArtistRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// Create functions creates artist.
func (r *ArtistRepository) Create(ctx context.Context, artist entities.Artist) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.InsertOne(ctx, artist)
	if err != nil {
		return err
	}
	return nil
}

// FindArtistByID finds artist by ID.
func (r *ArtistRepository) FindArtistByID(ctx context.Context, id primitive.ObjectID) (*entities.Artist, error) {
	var artist entities.Artist
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}
