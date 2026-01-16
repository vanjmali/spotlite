package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
)

// SongDto represents the data transfer object for a song entity.
type SongDto struct {
	Title         string   `json:"title" validate:"required,min=2"`
	Genre         string   `json:"genre" validate:"required,min=2"`
	LengthSeconds int      `json:"length_seconds" validate:"required,gt=0"`
	ArtistIds     []string `json:"artist_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
}

// SongListResponseDto represents the response payload when returning a paginated list of songs.
type SongListResponseDto = commondtos.ItemCollectionResponse[entities.Song]
