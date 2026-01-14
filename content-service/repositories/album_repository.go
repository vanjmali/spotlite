package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlbumRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *AlbumRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

func NewAlbumRepository(dbName string, collName string, c *mongo.Client) *AlbumRepository {
	r := AlbumRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *AlbumRepository) Create(ctx context.Context, album entities.Album) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, album)
	if err != nil {
		return err
	}
	return nil
}

func (r *AlbumRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Album, error) {
	c := r.getCollection()

	var album entities.Album

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&album); err != nil {
		return nil, err
	}
	return &album, nil
}
