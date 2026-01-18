package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

// Artist models an artist document stored in MongoDB.
type Artist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Genres      []string           `bson:"genres" json:"genres"`
	Description string             `bson:"description" json:"description"`
}
