package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PasswordRecoveryRepository provides data access for password recovery tokens.
type PasswordRecoveryRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *PasswordRecoveryRepository) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewPasswordRecoveryRepository constructs a PasswordRecoveryRepository for the given database and collection.
func NewPasswordRecoveryRepository(dbName string, collName string, c *mongo.Client) *PasswordRecoveryRepository {
	r := PasswordRecoveryRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// SaveRecoveryToken inserts a new password recovery token into the database.
func (r *PasswordRecoveryRepository) SaveRecoveryToken(ctx context.Context, token *entities.PasswordRecoveryToken) error {
	c := r.getCollection()

	if token.ID == primitive.NilObjectID {
		token.ID = primitive.NewObjectID()
	}

	_, err := c.InsertOne(ctx, token)
	return err
}

// FindValidToken retrieves a valid (non-expired, non-used) recovery token by its searchable hash.
// The tokenSearchable parameter should be a SHA256 hash of the plaintext token.
func (r *PasswordRecoveryRepository) FindValidToken(ctx context.Context, tokenSearchable string) (*entities.PasswordRecoveryToken, error) {
	c := r.getCollection()

	filter := bson.M{
		"token_searchable": tokenSearchable,
		"used_at":          nil,
		"expires_at":       bson.M{"$gt": time.Now()},
	}

	var token entities.PasswordRecoveryToken
	err := c.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	return &token, nil
}

// MarkTokenAsUsed marks a recovery token as used by setting the UsedAt timestamp.
func (r *PasswordRecoveryRepository) MarkTokenAsUsed(ctx context.Context, tokenID primitive.ObjectID) error {
	c := r.getCollection()

	filter := bson.M{"_id": tokenID}
	update := bson.M{
		"$set": bson.M{
			"used_at": time.Now(),
		},
	}

	res, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return ErrTokenExpired
	}

	return nil
}

// InvalidateAllTokensForUser marks all recovery tokens for a user as used,
// preventing them from being used again.
func (r *PasswordRecoveryRepository) InvalidateAllTokensForUser(ctx context.Context, userID primitive.ObjectID) error {
	c := r.getCollection()

	filter := bson.M{
		"user_id": userID,
		"used_at": nil, // Only invalidate unused tokens
	}

	update := bson.M{
		"$set": bson.M{
			"used_at": time.Now(),
		},
	}

	_, err := c.UpdateMany(ctx, filter, update)
	return err
}
