package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Title         string             `bson:"title"`
	Genre         string             `bson:"genre"`
	Artists       []Artist           `bson:"artists"`
	LengthSeconds int                `bson:"length_seconds"`
}
