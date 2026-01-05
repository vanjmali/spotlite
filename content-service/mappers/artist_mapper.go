package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToArtistEntity(a *dtos.ArtistDto) (*entities.Artist, error) {
	return &entities.Artist{
		ID:          primitive.NewObjectID(),
		Name:        a.Name,
		Genres:      a.Genres,
		Description: a.Description,
	}, nil
}
