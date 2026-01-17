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

// UpdateByID updates album by ID.
func (r *AlbumRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update map[string]any) (*entities.Album, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedAlbum entities.Album
	err := c.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedAlbum)
	if err != nil {
		return nil, err
	}

	return &updatedAlbum, nil
}

// DeleteByID deletes album by ID.
func (r *AlbumRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	res, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return res, err
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

	var albums []entities.Album = make([]entities.Album, 0) // Initialize with empty slice instead of nil
	if err := cur.All(ctx, &albums); err != nil {
		return nil, 0, err
	}

	return albums, total, nil
}
