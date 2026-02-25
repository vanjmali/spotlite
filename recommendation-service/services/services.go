package services

import (
	"context"

	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
)

// GraphRelationRepository defines methods needed for graph recommendations
type GraphRelationRepository interface {
	GetRecommendedSongsForUser(ctx context.Context, userID string, limit int) ([]string, error)
	GetCollaborativeRecommendations(ctx context.Context, userID string, limit int) ([]string, error)
	GetHighlyRatedSongs(ctx context.Context, minRating float64, limit int) ([]string, error)
	GetSongRatingStats(ctx context.Context, songID string) (avgRating float64, ratingCount int64, err error)
	GetSongArtists(ctx context.Context, songID string) ([]string, error)
}

// SongNodeRepository defines methods for accessing song nodes
type SongNodeRepository interface {
	Get(ctx context.Context, songID string) (*entities.SongNode, error)
}

type Services struct {
	userNodeRepository   *repositories.UserNodeRepository
	songNodeRepository   SongNodeRepository
	artistNodeRepository *repositories.ArtistNodeRepository
	genreNodeRepository  *repositories.GenreNodeRepository
	albumNodeRepository  *repositories.AlbumNodeRepository
	relationRepository   GraphRelationRepository
}

func NewServices(
	ur *repositories.UserNodeRepository,
	sr SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
	ab *repositories.AlbumNodeRepository,
	rr GraphRelationRepository,
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
