package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AlbumRepository provides data access helpers for album documents.
type AlbumRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *AlbumRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewAlbumRepository constructs a AlbumRepository for the given database and collection.
func NewAlbumRepository(dbName string, collName string, c *mongo.Client) *AlbumRepository {
	r := AlbumRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// Create func, inserts a new album into the database.
func (r *AlbumRepository) Create(ctx context.Context, album entities.Album) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, album)
	if err != nil {
		return err
	}
	return nil
}

// FindByID finds album by ID.
func (r *AlbumRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Album, error) {
	c := r.getCollection()

	var album entities.Album

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&album); err != nil {
		return nil, err
	}
	return &album, nil
}

// FindAll func, finds all albums matching the filter with pagination.
func (r *AlbumRepository) FindAll(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Album, int64, error) {
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

	var albums []entities.Album
	if err := cur.All(ctx, &albums); err != nil {
		return nil, 0, err
	}

	return albums, total, nil
}
