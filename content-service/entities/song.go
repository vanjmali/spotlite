package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Song models a song document stored in MongoDB.
type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Title         string             `bson:"title"`
	Genre         string             `bson:"genre"`
	LengthSeconds int                `bson:"length_seconds"`
	Artists       []Artist           `bson:"artists"`
}
