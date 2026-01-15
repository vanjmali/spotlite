package dtos

import (
	"time"

	"github.com/vanjmali/spotlite/content/entities"
)

// CreateAlbumDto represents the payload required to create a new album.
type CreateAlbumDto struct {
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"release_date"`
	Genres      []string  `json:"genres"`
	SongIds     []string  `json:"song_ids"`
	ArtistIds   []string  `json:"artist_ids"`
}

// AlbumQueryDto represents the query parameters for filtering and paginating album results.
type AlbumQueryDto struct {
	Page     int
	Size     int
	Name    string
	Genres    string
	ArtistID string
}

// AlbumListResponseDto represents the response payload when returning a paginated list of albums.
type AlbumListResponseDto struct {
	Items []entities.Album `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}
