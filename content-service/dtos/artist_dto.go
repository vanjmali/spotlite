package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArtistDto represents the data transfer object for an artist entity.
type ArtistDto struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" validate:"required,min=2"`
	Genres      []string           `bson:"genres" json:"genres" validate:"required,min=1,dive,required,min=2,max=30"`
	Description string             `bson:"description" json:"description" validate:"required,min=2"`
}

// UpdateArtistDto represents the payload for updating an existing artist. Pointers allow partial updates.
type UpdateArtistDto struct {
	Name        *string   `json:"name" validate:"omitempty,min=2"`
	Genres      *[]string `json:"genres" validate:"omitempty,min=1,dive,required,min=2,max=30"`
	Description *string   `json:"description" validate:"omitempty,min=2"`
}

// ArtistListResponseDto represents the response payload when returning a paginated list of artists.
type ArtistListResponseDto = commondtos.ItemCollectionResponse[entities.Artist]
