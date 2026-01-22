package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToArtistEntity converts an ArtistDto to an Artist entity.
func ToArtistEntity(a *dtos.ArtistDto, genres []entities.Genre) (*entities.Artist, error) {

	return &entities.Artist{
		ID:          primitive.NewObjectID(),
		Name:        a.Name,
		Genres:      genres,
		Description: a.Description,
	}, nil
}
