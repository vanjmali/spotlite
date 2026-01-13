package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *SongRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Song, error) {
	c := r.getCollection()

	var song entities.Song

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&song); err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) DeleteById(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	res, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	return res, err

}
