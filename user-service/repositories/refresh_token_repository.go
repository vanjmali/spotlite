package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefreshTokenRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

const RefreshTokensColl = "refresh_tokens"

func (r *RefreshTokenRepository) refreshColl() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(RefreshTokensColl)
}

func NewRefreshTokenRepository(dbName string, collName string, c *mongo.Client) *RefreshTokenRepository {
	r := RefreshTokenRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *RefreshTokenRepository) EnsureRefreshIndexes(ctx context.Context) error {
	c := r.refreshColl()
	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	})
	return err
}

func (r *RefreshTokenRepository) InsertToken(ctx context.Context, rt entities.RefreshToken) (primitive.ObjectID, error) {
	res, err := r.refreshColl().InsertOne(ctx, rt)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *RefreshTokenRepository) FindActiveByHash(ctx context.Context, hash string) (*entities.RefreshToken, error) {
	var rt entities.RefreshToken
	err := r.refreshColl().FindOne(ctx, bson.M{
		"token_hash": hash,
		"revoked_at": bson.M{"$exists": false},
	}).Decode(&rt)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, mongo.ErrNoDocuments
	}

	return &rt, err
}

func (r *RefreshTokenRepository) RevokeByID(ctx context.Context, id primitive.ObjectID, when time.Time, replacedBy primitive.ObjectID) error {
	_, err := r.refreshColl().UpdateByID(ctx, id, bson.M{
		"$set": bson.M{
			"revoked_at":     when,
			"replaced_by_id": replacedBy,
			"last_used_at":   time.Now(),
		},
	})
	return err
}
