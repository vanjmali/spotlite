package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToGenreEntity(a *dtos.GenreDto) (*entities.Genre, error) {
	return &entities.Genre{
		ID:   primitive.NewObjectID(),
		Name: a.Name,
	}, nil
}
