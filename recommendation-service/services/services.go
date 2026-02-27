package services

import (
	"context"

	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
)

// GraphRelationRepository defines methods needed for graph recommendations.
type GraphRelationRepository interface {
	SaveSongWithGenres(ctx context.Context, e events.SongCreationPayload) error
	CreateGenreSubscription(ctx context.Context, e events.GenreSubscriptionEventPayload) error
	CreateRating(ctx context.Context, e events.SongRatingPayload) error
	UpdateSongWithGenres(ctx context.Context, e events.SongUpdatePayload) error
	UpdateGenre(ctx context.Context, e events.EntityUpdatedEventPayload) error
}

// SongNodeRepository defines methods for accessing song nodes.
type SongNodeRepository interface {
	Get(ctx context.Context, songID string) (*entities.SongNode, error)
}

func NewServices(
	ur *repositories.UserNodeRepository,
	gr *repositories.GenreNodeRepository,
	rr GraphRelationRepository,
) *Repositories {
	return &Repositories{
		userNodeRepository:  ur,
		genreNodeRepository: gr,
		relationRepository:  rr,
	}
}
