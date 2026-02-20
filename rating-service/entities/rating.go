package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Rating struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SongID   primitive.ObjectID `bson:"song_id" json:"songId"`
	UserID   primitive.ObjectID `bson:"user_id" json:"userId"`
	Username string             `bson:"username" json:"username"`
	Value    int                `bson:"value" json:"value"`
	IsEdited bool               `bson:"is_edited" json:"isEdited"`
}
