package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Album models an album document stored in MongoDB.
type Album struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	ReleaseDate string             `bson:"release_date" json:"release_date"` // Date string in YYYY-MM-DD format
	Genres      []Genre            `bson:"genres" json:"genres"`
	Songs       []Song             `bson:"songs" json:"songs"`
	Artists     []Artist           `bson:"artists" json:"artists"`
}
