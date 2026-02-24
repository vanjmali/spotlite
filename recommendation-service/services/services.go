package services

import (
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
)

type Services struct {
	userNodeRepository   *repositories.UserNodeRepository
	songNodeRepository   *repositories.SongNodeRepository
	artistNodeRepository *repositories.ArtistNodeRepository
	genreNodeRepository  *repositories.GenreNodeRepository
	albumNodeRepository  *repositories.AlbumNodeRepository
	relationRepository   *repositories.GraphRelationRepository
}

func NewServices(
	ur *repositories.UserNodeRepository,
	sr *repositories.SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
	ab *repositories.AlbumNodeRepository,
	rr *repositories.GraphRelationRepository,
) *Services {
	return &Services{
		userNodeRepository:   ur,
		songNodeRepository:   sr,
		artistNodeRepository: ar,
		genreNodeRepository:  gr,
		albumNodeRepository:  ab,
		relationRepository:   rr,
	}
}
