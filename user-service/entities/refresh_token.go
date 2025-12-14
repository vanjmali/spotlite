package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	UserID       primitive.ObjectID  `bson:"user_id"`
	TokenHash    string              `bson:"token_hash"`
	ExpiresAt    time.Time           `bson:"expires_at"`
	RevokedAt    *time.Time          `bson:"revoked_at,omitempty"`
	LastUsedAt   *time.Time          `bson:"last_used_at,omitempty"`
	ReplacedByID *primitive.ObjectID `bson:"replaced_by_id,omitempty"`
}
