package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GenreDto struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name" validate:"required"`
}

// UpdateGenreDto represents the payload for updating an existing genre.
type UpdateGenreDto struct {
	Name *string `json:"name" validate:"omitempty,min=2"`
}

// GenreListResponseDto represents the response payload when returning a paginated list of genres.
type GenreListResponseDto = commondtos.ItemCollectionResponse[entities.Genre]
