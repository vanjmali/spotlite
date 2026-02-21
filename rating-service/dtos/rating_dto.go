package dtos

import "github.com/vanjmali/spotlite/rating-service/entities"

type CreateRatingDto struct {
	SongID string `json:"song_id" validate:"required"`
	Value  int    `json:"value" validate:"required,min=1,max=5"`
}

type RatingResponseDto struct {
	Items      []*entities.Rating `json:"items"`
	NextCursor string             `json:"nextCursor,omitempty"`
}
