package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type SongRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *SongRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewRepository constructs a ArtistRepository for the given database and collection.
func NewSongRepository(dbName string, collName string, c *mongo.Client) *SongRepository {
	r := SongRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *SongRepository) Create(ctx context.Context, song entities.Song) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, song)
	if err != nil {
		return err
	}
	return nil
}
