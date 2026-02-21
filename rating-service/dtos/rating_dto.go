package dtos

type CreateRatingDto struct {
	SongID string `json:"song_id" validate:"required"`
	Value  int    `json:"value" validate:"required,min=1,max=5"`
}
