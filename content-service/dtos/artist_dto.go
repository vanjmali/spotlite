package dtos

import (
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArtistDto struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Genres      []string           `bson:"genres" json:"genres"`
	Description string             `bson:"description" json:"description"`
}

type UpdateArtistDto struct {
	Name        *string   `json:"name"`
	Genres      *[]string `json:"genres"`
	Description *string   `json:"description"`
}

type ArtistQueryDto struct {
	Page  int
	Size  int
	Name  string
	Genre string
}

type ArtistListResponseDto struct {
	Items []entities.Artist `json:"items"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
	Total int64             `json:"total"`
}
