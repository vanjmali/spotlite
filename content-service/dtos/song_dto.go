package dtos

type SongDto struct {
	Title         string   `json:"title"`
	Genre         string   `json:"genre"`
	LengthSeconds int      `json:"length_seconds"`
	ArtistIds     []string `json:"artist_ids"`
}
