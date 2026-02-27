package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Error types for recommendation service.
var (
	ErrGraphDatabaseUnavailable = errors.New("graph database is currently unavailable")
	ErrLimitOutOfRange          = errors.New("limit must be between 1 and 50")
	ErrInvalidUserID            = errors.New("invalid user ID format")
)

type Repositories struct {
	userNodeRepository  *repositories.UserNodeRepository
	genreNodeRepository *repositories.GenreNodeRepository
	relationRepository  GraphRelationRepository
}

// RecommendationService provides recommendation-related business logic.
type RecommendationService struct {
	r  *Repositories
	tr trace.Tracer
}

// NewRecommendationService constructs a RecommendationService.
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

func (rs *RecommendationService) CreateGenre(g events.GenreCreationPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.user.create")
	defer createSpan.End()

	gn := entities.GenreNode{GenreID: g.GenreID, Name: g.GenreName}

	err := rs.r.genreNodeRepository.Create(createCtx, gn)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while creating genre: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) CreateSong(e events.SongCreationPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.song.create")
	defer createSpan.End()

	err := rs.r.relationRepository.SaveSongWithGenres(createCtx, e)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while creating song: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) CreateSubscription(e events.GenreSubscriptionEventPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.subscription.create")
	defer createSpan.End()

	err := rs.r.relationRepository.CreateGenreSubscription(createCtx, e)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while creating subscription relationship: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) CreateRating(e events.SongRatingPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.rating.create")
	defer createSpan.End()

	err := rs.r.relationRepository.CreateRating(createCtx, e)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while creating rating relationship: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) UpdateSong(e events.SongUpdatePayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.song.update")
	defer createSpan.End()

	err := rs.r.relationRepository.UpdateSongWithGenres(createCtx, e)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while updating song node and it's relationships: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) UpdateGenre(e events.EntityUpdatedEventPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.genre.update")
	defer createSpan.End()

	err := rs.r.relationRepository.UpdateGenre(createCtx, e)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while updating genre node and it's relationships: %v", err)
		return err
	}

	return nil
}
