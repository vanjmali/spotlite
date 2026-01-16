package dtos

import (
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/types"
)

// CreateAlbumDto represents the payload required to create a new album.
type CreateAlbumDto struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	ReleaseDate types.Date `json:"release_date" validate:"required,notzerodate"`
	Genres      []string   `json:"genres" validate:"required,min=1,dive,required,min=2,max=30"`
	SongIds     []string   `json:"song_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
	ArtistIds   []string   `json:"artist_ids" validate:"required,min=1,dive,required,len=24,hexadecimal"`
}

// AlbumListResponseDto represents the response payload when returning a paginated list of albums.
type AlbumListResponseDto struct {
	Items []entities.Album `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}
