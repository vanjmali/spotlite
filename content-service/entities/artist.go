package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

// Artist models an artist document stored in MongoDB.
type Artist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Genres      []string           `bson:"genres"`
	Description string             `bson:"description"`
}
