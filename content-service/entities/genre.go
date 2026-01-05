package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Genre struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name,omitempty"`
}
