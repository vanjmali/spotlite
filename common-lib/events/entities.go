package events

import (
	"time"
)

const (
	CONTENT_STREAM = "CONTENT"

	SUBJECT_ARTIST_CREATED = "content.created.artist"
	ARTIST_DURABLE         = "ARTIST_CREATOR"

	SUBJECT_ALBUM_CREATED = "content.created.album"
	ALBUM_DURABLE         = "ALBUM_CREATOR"
)

type ArtistEventPayload struct {
	GenreIds   []string  `json:"genre_ids"`
	CreatedAt  time.Time `json:"created_at"`
	ArtistID   string    `json:"artist_id"`
	ArtistName string    `json:"artist_name"`
}
