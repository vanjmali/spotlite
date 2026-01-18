package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Song models a song document stored in MongoDB.
type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Genre         string             `bson:"genre" json:"genre"`
	LengthSeconds int                `bson:"length_seconds" json:"lengthSeconds"`
	Artists       []Artist           `bson:"artists" json:"artists"`
}
