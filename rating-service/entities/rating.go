package entities

import (
	"time"

	"github.com/vanjmali/spotlite/common-lib/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Rating represents a user's rating for a song stored in MongoDB.
type Rating struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SongID    primitive.ObjectID `bson:"song_id" json:"song_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Value     int                `bson:"value" json:"value"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	IsEdited  bool               `bson:"is_edited" json:"is_edited"`
	Status    types.EntityStatus `bson:"status" json:"status"`
}
