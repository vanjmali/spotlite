package mappers

import (
	"github.com/vanjmali/spotlite/common-lib/types"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToSongEntity converts a SongDto to a Song entity with associated genres and artists.
func ToSongEntity(a *dtos.SongDto, genres []entities.Genre, artists []entities.Artist) (*entities.Song, error) {
	return &entities.Song{
		ID:      primitive.NewObjectID(),
		Title:   a.Title,
		Genres:  genres,
		Artists: artists,
		Status:  types.StatusActive,
	}, nil
}
