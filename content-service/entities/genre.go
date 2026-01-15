package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

// Genre models a genre document stored in MongoDB.
type Genre struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name,omitempty"`
}
