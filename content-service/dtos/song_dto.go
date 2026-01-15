package dtos

import "github.com/vanjmali/spotlite/content/entities"

// SongDto represents the data transfer object for a song entity.
type SongDto struct {
	Title         string   `json:"title"`
	Genre         string   `json:"genre"`
	LengthSeconds int      `json:"length_seconds"`
	ArtistIds     []string `json:"artist_ids"`
}

// SongQueryDto represents the query parameters for filtering and paginating song results.
type SongQueryDto struct {
	Page     int
	Size     int
	Title    string
	Genre    string
	ArtistID string
}

// SongListResponseDto represents the response payload when returning a paginated list of songs.
type SongListResponseDto struct {
	Items []entities.Song `json:"items"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Total int64           `json:"total"`
}
