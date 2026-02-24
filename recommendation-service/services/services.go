package services

import (
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
)

type Services struct {
	userNodeRepository   *repositories.UserNodeRepository
	songNodeRepository   *repositories.SongNodeRepository
	artistNodeRepository *repositories.ArtistNodeRepository
	genreNodeRepository  *repositories.GenreNodeRepository
}

func NewServices(
	ur *repositories.UserNodeRepository,
	sr *repositories.SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
) *Services {
	return &Services{
		userNodeRepository:   ur,
		songNodeRepository:   sr,
		artistNodeRepository: ar,
		genreNodeRepository:  gr,
	}
}
