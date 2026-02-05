package dtos

import "github.com/vanjmali/spotlite/content/entities"

// GlobalSearchResponseDto represents the response for a global search.
type GlobalSearchResponseDto struct {
	Genres  []entities.Genre  `json:"genres"`
	Albums  []entities.Album  `json:"albums"`
	Songs   []entities.Song   `json:"songs"`
	Artists []entities.Artist `json:"artists"`
}
