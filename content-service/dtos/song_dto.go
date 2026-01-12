package dtos

import "github.com/vanjmali/spotlite/content/entities"

type CreateSongDto struct {
	Title         string            `json:"title"`
	Genre         string            `json:"genre"`
	Artists       []entities.Artist `json:"artists,omitempty"`
	LengthSeconds int               `json:"length_seconds"`
}
