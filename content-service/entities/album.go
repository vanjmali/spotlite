package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Album models an album document stored in MongoDB.
type Album struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	ReleaseDate time.Time          `bson:"release_date"`
	Genres      []string           `bson:"genres"`
	Songs       []Song             `bson:"songs"`
	Artists     []Artist           `bson:"artists"`
}
