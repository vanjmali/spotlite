package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/rating-service/entities"
)

// CreateRatingDto represents the payload required to create a new rating.
type CreateRatingDto struct {
	SongID string `json:"song_id" validate:"required"`
	Value  int    `json:"value" validate:"required,min=1,max=5"`
}

// UpdateRatingDto represents the payload for updating an existing rating. The Value field is a pointer to allow partial updates.
type UpdateRatingDto struct {
	Value *int `json:"value" validate:"required,min=1,max=5"`
}

// RatingResponseDto represents the response payload when returning a paginated list of ratings.
type RatingResponseDto struct {
	Items      []*entities.Rating `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

// SongRatingSummary represents the average rating and total count of ratings for a song.
type SongRatingSummary struct {
	Avg   float64 `bson:"avg" json:"avg"`
	Count int64   `bson:"count" json:"count"`
}

// RatingListResponseDto represents the response payload when returning a paginated list of ratings.
type RatingListResponseDto = commondtos.ItemCollectionResponse[entities.Rating]
