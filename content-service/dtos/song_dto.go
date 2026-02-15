package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
)

// SongDto represents the data transfer object for a song entity.
type SongDto struct {
	Title     string   `json:"title" validate:"required,min=2"`
	AlbumId   string   `json:"album_id" validate:"required,len=24,hexadecimal"`
	GenreIds  []string `json:"genre_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
	ArtistIds []string `json:"artist_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
}

// UpdateSongDto represents the payload for updating an existing song. Pointers allow partial updates.
type UpdateSongDto struct {
	Title     *string   `json:"title" validate:"omitempty,required,min=2"`
	GenreIds  *[]string `json:"genre_ids" validate:"omitempty,min=1,dive,required,len=24,hexadecimal"`
	ArtistIds *[]string `json:"artist_ids" validate:"omitempty,min=1,dive,required,len=24,hexadecimal"`
}

// SongListResponseDto represents the response payload when returning a paginated list of songs.
type SongListResponseDto = commondtos.ItemCollectionResponse[entities.Song]
