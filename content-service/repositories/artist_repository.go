package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type ArtistRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func NewRepository(dbName string, collName string, c *mongo.Client) *ArtistRepository {
	r := ArtistRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *ArtistRepository) Create(ctx context.Context, artist entities.Artist) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.InsertOne(ctx, artist)
	if err != nil {
		return err
	}
	return nil
}
