package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/content/entities"
)

// CreateAlbumDto represents the payload required to create a new album.
type CreateAlbumDto struct {
	Title       string   `json:"title" validate:"required,min=2,max=100"`
	ReleaseDate string   `json:"release_date" validate:"required,len=10"` // Date string in YYYY-MM-DD format
	GenreIds    []string `json:"genre_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
	SongIds     []string `json:"song_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
	ArtistIds   []string `json:"artist_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
}

// UpdateAlbumDto represents the payload for updating an existing album. Pointers allow partial updates.
type UpdateAlbumDto struct {
	Title       *string   `json:"title" validate:"omitempty,min=2,max=100"`
	ReleaseDate *string   `json:"release_date" validate:"omitempty,len=10"` // Date string in YYYY-MM-DD format
	GenreIds    *[]string `json:"genre_ids" validate:"omitempty,min=1,dive,required,len=24,hexadecimal"`
	ArtistIds   *[]string `json:"artist_ids" validate:"omitempty,min=1,dive,required,len=24,hexadecimal"`
}

// AddAlbumSongsDto represents the payload for adding songs to an album.
type AddAlbumSongsDto struct {
	Ids []string `json:"ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
}

// AlbumListResponseDto represents the response payload when returning a paginated list of albums.
type AlbumListResponseDto = commondtos.ItemCollectionResponse[entities.Album]
