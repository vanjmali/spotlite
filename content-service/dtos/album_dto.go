package dtos

import (
	"time"

	"github.com/vanjmali/spotlite/content/entities"
)

type CreateAlbumDto struct {
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"release_date"`
	Genres      []string  `json:"genres"`
	SongIds     []string  `json:"song_ids"`
	ArtistIds   []string  `json:"artist_ids"`
}

type AlbumQueryDto struct {
	Page     int
	Size     int
	Title    string
	Genre    string
	ArtistID string
}

type AlbumListResponseDto struct {
	Items []entities.Album `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}
