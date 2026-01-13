package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Artist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Genres      []string           `bson:"genres"`
	Description string             `bson:"description"`
}

type EmbeddedArtist struct {
	ArtistID    primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Genres      []string           `bson:"genres"`
	Description string             `bson:"description"`
}
