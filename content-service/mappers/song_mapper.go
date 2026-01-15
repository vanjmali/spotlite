package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToSongEntity converts a SongDto to a Song entity with associated artists.
func ToSongEntity(a *dtos.SongDto, artists []entities.Artist) (*entities.Song, error) {

	return &entities.Song{
		ID:            primitive.NewObjectID(),
		Title:         a.Title,
		Genre:         a.Genre,
		LengthSeconds: a.LengthSeconds,
		Artists:       artists,
	}, nil
}
