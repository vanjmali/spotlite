package repositories

import (
	"context"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SongRepository provides data access helpers for song documents.
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

// Create func, inserts a new song into the database.
func (r *SongRepository) Create(ctx context.Context, song entities.Song) (primitive.ObjectID, error) {
	c := r.getCollection()

	res, err := c.InsertOne(ctx, song)
	if err != nil {
		return primitive.NilObjectID, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, mongo.ErrNilDocument
	}

	return oid, nil
}

// FindByID finds song by ID.
func (r *SongRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Song, error) {
	c := r.getCollection()

	var song entities.Song

	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&song); err != nil {
		return nil, err
	}
	return &song, nil
}

// UpdateByID updates song by ID.
func (r *SongRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update map[string]any) (*entities.Song, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedSong entities.Song
	err := c.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedSong)
	if err != nil {
		return nil, err
	}

	return &updatedSong, nil
}

// DeleteByID deletes song by ID.
func (r *SongRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := r.getCollection()

	filter := bson.M{"_id": id}
	res, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	return res, err
}

// FindAll func, finds all songs matching the filter with pagination.
func (r *SongRepository) FindAll(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Song, int64, error) {
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

	songs := make([]entities.Song, 0)
	if err = cur.All(ctx, &songs); err != nil {
		return nil, 0, err
	}

	return songs, total, nil
}

func (r *SongRepository) UpdateAudioByID(ctx context.Context, id primitive.ObjectID, audioPath string, size int64, mime string) (*entities.Song, error) {
	return r.UpdateByID(ctx, id, map[string]any{
		"audio_path":      audioPath,
		"audio_size":      size,
		"audio_mime_type": mime,
	})
}
