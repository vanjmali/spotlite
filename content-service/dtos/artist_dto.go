package dtos

type ArtistDto struct {
	Name        string   `json:"name"`
	Genres      []string `json:"genres"`
	Description string   `json:"description"`
}

type UpdateArtistDto struct {
	Name        *string   `json:"name"`
	Genres      *[]string `json:"genres"`
	Description *string   `json:"description"`
}
