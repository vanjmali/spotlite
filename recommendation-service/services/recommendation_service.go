package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Error types for recommendation service
var (
	ErrGraphDatabaseUnavailable = errors.New("graph database is currently unavailable")
	ErrLimitOutOfRange          = errors.New("limit must be between 1 and 50")
	ErrInvalidUserID            = errors.New("invalid user ID format")
)

type Repositories struct {
	userNodeRepository   *repositories.UserNodeRepository
	songNodeRepository   SongNodeRepository
	artistNodeRepository *repositories.ArtistNodeRepository
	genreNodeRepository  *repositories.GenreNodeRepository
	albumNodeRepository  *repositories.AlbumNodeRepository
	relationRepository   GraphRelationRepository
}

// RecommendationService provides recommendation-related business logic
type RecommendationService struct {
	r  *Repositories
	tr trace.Tracer
}

// NewRecommendationService constructs a RecommendationService
func NewRecommendationService(r *Repositories) *RecommendationService {
	return &RecommendationService{
		r:  r,
		tr: otel.Tracer("recommendation-service/recommendation-service"),
	}
}

func (rs *RecommendationService) CreateUser(u events.UserRegistrationPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.user.create")
	defer createSpan.End()

	un := entities.UserNode{UserID: u.UserID, Username: u.Username}

	err := rs.r.userNodeRepository.Create(createCtx, un)
	if err != nil {
		return err
	}

	return nil
}
