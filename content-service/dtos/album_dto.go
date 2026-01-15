package dtos

import (
	"time"

	"github.com/vanjmali/spotlite/content/entities"
)

// CreateAlbumDto represents the payload required to create a new album.
type CreateAlbumDto struct {
	Name        string    `json:"name" validate:"required"`
	ReleaseDate time.Time `json:"release_date" validate:"required"`
	Genres      []string  `json:"genres" validate:"required"`
	SongIds     []string  `json:"song_ids" validate:"required"`
	ArtistIds   []string  `json:"artist_ids" validate:"required"`
}

// AlbumListResponseDto represents the response payload when returning a paginated list of albums.
type AlbumListResponseDto struct {
	Items []entities.Album `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}
