package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArtistRepository provides data access helpers for artist documents.
type ArtistRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *ArtistRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewRepository constructs a ArtistRepository for the given database and collection.
func NewArtistRepository(dbName string, collName string, c *mongo.Client) *ArtistRepository {
	r := ArtistRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// Create functions creates artist.
func (r *ArtistRepository) Create(ctx context.Context, artist entities.Artist) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, artist)
	if err != nil {
		return err
	}
	return nil
}

// FindByID finds artist by ID.
func (r *ArtistRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Artist, error) {
	var artist entities.Artist
	c := r.getCollection()

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

// UpdateByID updates artist by ID.
func (r *ArtistRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update map[string]any) (*entities.Artist, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedArtist entities.Artist
	err := c.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedArtist)
	if err != nil {
		return nil, err
	}

	return &updatedArtist, nil
}

// DeleteByID deletes artist by ID.
func (r *ArtistRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}

	res, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return res, err
}

// FindAll func, finds all artists matching the filter with pagination.
func (r *ArtistRepository) FindAll(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Artist, int64, error) {
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

	artists := make([]entities.Artist, 0)
	if err := cur.All(ctx, &artists); err != nil {
		return nil, 0, err
	}

	return artists, total, nil
}

// Exists func, checks if artist with the given ID exists.
func (r *ArtistRepository) Exists(ctx context.Context, artistID primitive.ObjectID) (bool, error) {
	cnt, err := r.getCollection().CountDocuments(
		ctx,
		bson.M{"_id": artistID},
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}
