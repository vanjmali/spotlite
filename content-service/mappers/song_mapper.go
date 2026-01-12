package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToSongEntity(a *dtos.CreateSongDto) (*entities.Song, error) {
	return &entities.Song{
		ID:            primitive.NewObjectID(),
		Title:         a.Title,
		Genre:         a.Genre,
		Artists:       a.Artists,
		LengthSeconds: a.LengthSeconds,
	}, nil
}
