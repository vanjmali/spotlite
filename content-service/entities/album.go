package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Album struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	Name        string               `bson:"name"`
	ReleaseDate time.Time            `bson:"release_date"`
	Genres      []string             `bson:"genres"`
	SongIds     []primitive.ObjectID `bson:"song_ids"`
	ArtistIds   []primitive.ObjectID `bson:"artists"`
}
