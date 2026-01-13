package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Title         string             `bson:"title"`
	Genre         string             `bson:"genre"`
	LengthSeconds int                `bson:"length_seconds"`
	Artists       []EmbeddedArtist   `bson:"artists"`
}
