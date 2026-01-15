package dtos

import (
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArtistDto represents the data transfer object for an artist entity.
type ArtistDto struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Genres      []string           `bson:"genres" json:"genres"`
	Description string             `bson:"description" json:"description"`
}

// UpdateArtistDto represents the payload for updating an existing artist. Pointers allow partial updates.
type UpdateArtistDto struct {
	Name        *string   `json:"name"`
	Genres      *[]string `json:"genres"`
	Description *string   `json:"description"`
}

// ArtistQueryDto represents the query parameters for filtering and paginating artist results.
type ArtistQueryDto struct {
	Page  int
	Size  int
	Name  string
	Genre string
}

// ArtistListResponseDto represents the response payload when returning a paginated list of artists.
type ArtistListResponseDto struct {
	Items []entities.Artist `json:"items"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
	Total int64             `json:"total"`
}
