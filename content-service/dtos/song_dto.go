package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
)

// SongDto represents the data transfer object for a song entity.
type SongDto struct {
	Title         string   `json:"title" validate:"required"`
	Genre         string   `json:"genre" validate:"required"`
	LengthSeconds int      `json:"length_seconds" validate:"required,min=1"`
	ArtistIds     []string `json:"artist_ids" validate:"required"`
}

// SongListResponseDto represents the response payload when returning a paginated list of songs.
type SongListResponseDto = commondtos.ItemCollectionResponse[entities.Song]
