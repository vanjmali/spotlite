package dtos

import "go.mongodb.org/mongo-driver/bson/primitive"

type GenreDto struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name" validate:"required"`
}
