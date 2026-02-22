package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/rating-service/entities"
)

type CreateRatingDto struct {
	SongID string `json:"song_id" validate:"required"`
	Value  int    `json:"value" validate:"required,min=1,max=5"`
}

type UpdateRatingDto struct {
	Value *int `json:"value" validate:"required,min=1,max=5"`
}

type RatingResponseDto struct {
	Items      []*entities.Rating `json:"items"`
	NextCursor string             `json:"nextCursor,omitempty"`
}

type RatingListResponseDto = commondtos.ItemCollectionResponse[entities.Rating]
