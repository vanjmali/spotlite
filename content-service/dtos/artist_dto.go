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

type ArtistQueryDto struct {
	Page  int
	Size  int
	Name  string
	Genre string
}

type ArtistListResponseDto struct {
	Items []ArtistDto `json:"items"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int64       `json:"total"`
}
