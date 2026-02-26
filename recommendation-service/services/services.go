package services

import (
	"context"

	"github.com/vanjmali/spotlite/common-lib/events"
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
	SaveSongWithGenres(ctx context.Context, e events.SongCreationPayload) error
}

// SongNodeRepository defines methods for accessing song nodes
type SongNodeRepository interface {
	Get(ctx context.Context, songID string) (*entities.SongNode, error)
}

func NewServices(
	ur *repositories.UserNodeRepository,
	sr SongNodeRepository,
	gr *repositories.GenreNodeRepository,
	rr GraphRelationRepository,
) *Repositories {
	return &Repositories{
		userNodeRepository:  ur,
		songNodeRepository:  sr,
		genreNodeRepository: gr,
		relationRepository:  rr,
	}
}
