package dtos

import "github.com/vanjmali/spotlite/content/entities"

type SongDto struct {
	Title         string   `json:"title"`
	Genre         string   `json:"genre"`
	LengthSeconds int      `json:"length_seconds"`
	ArtistIds     []string `json:"artist_ids"`
}

type SongQueryDto struct {
	Page     int
	Size     int
	Title    string
	Genre    string
	ArtistID string
}

type SongListResponseDto struct {
	Items []entities.Song `json:"items"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Total int64           `json:"total"`
}
