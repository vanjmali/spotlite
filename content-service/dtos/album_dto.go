package dtos

import (
	"time"
)

type CreateAlbumDto struct {
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"release_date"`
	Genres      []string  `json:"genres"`
	SongIds     []string  `json:"song_ids,omitempty"`
	ArtistIds   []string  `json:"artist_ids"`
}
