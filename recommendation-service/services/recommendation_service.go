package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Error types for recommendation service.
var (
	ErrGraphDatabaseUnavailable = errors.New("graph database is currently unavailable")
	ErrLimitOutOfRange          = errors.New("limit must be between 1 and 50")
	ErrInvalidUserID            = errors.New("invalid user ID format")
	ErrObjectIdCastFailed       = errors.New("failed to convert hex to objectId")
)

type GraphRelationRepository interface {
	SaveSongWithGenres(ctx context.Context, sn entities.SongNode) error
	CreateGenreSubscription(ctx context.Context, gs entities.GenreSubscription) error
	CreateRating(ctx context.Context, sr entities.SongRating) error
	UpdateSongWithGenres(ctx context.Context, sn entities.SongNode) error
	UpdateGenre(ctx context.Context, gn entities.GenreNode) error
	FindSubscriptionBasedRecommendations(ctx context.Context, userID string) ([]*entities.SongRecommendation, error)
	FindLikeBasedRecommendation(ctx context.Context, userID string) ([]*entities.SongRecommendation, error)
	DeleteSong(ctx context.Context, songID string) error
}
type GenreNodeRepository interface {
	Create(ctx context.Context, genre entities.GenreNode) error
}

type UserNodeRepository interface {
	Create(ctx context.Context, user entities.UserNode) error
}

func NewServices(
	ur UserNodeRepository,
	gr GenreNodeRepository,
	rr GraphRelationRepository,
) *Repositories {
	return &Repositories{
		ur: ur,
		gr: gr,
		rr: rr,
	}
}

type Repositories struct {
	ur UserNodeRepository
	gr GenreNodeRepository
	rr GraphRelationRepository
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

	err := rs.r.ur.Create(createCtx, un)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RecommendationService) CreateGenre(g events.GenreCreationPayload, ctx context.Context) error {
	createCtx, createSpan := rs.tr.Start(ctx, "recommendation.user.create")
	defer createSpan.End()

	gn := entities.GenreNode{GenreID: g.GenreID, Name: g.GenreName}

	err := rs.r.gr.Create(createCtx, gn)
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

	sn := entities.SongNode{SongID: e.SongID, Title: e.SongTitle, Duration: e.Duration, GenreIDs: e.GenreIDs, Artists: e.ArtistNames}

	err := rs.r.rr.SaveSongWithGenres(createCtx, sn)
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

	gs := entities.GenreSubscription{GenreID: e.GenreID, UserID: e.UserID}

	err := rs.r.rr.CreateGenreSubscription(createCtx, gs)
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

	sr := entities.SongRating{SongID: e.SongID, UserID: e.UserID, Value: e.Value}

	err := rs.r.rr.CreateRating(createCtx, sr)
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

	sn := entities.SongNode{SongID: e.SongID, Title: e.SongTitle, Duration: e.Duration, GenreIDs: e.GenreIDs, Artists: e.ArtistNames}

	err := rs.r.rr.UpdateSongWithGenres(createCtx, sn)
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

	gn := entities.GenreNode{GenreID: e.EntityID, Name: e.EntityName}

	err := rs.r.rr.UpdateGenre(createCtx, gn)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(createCtx, "critical: an error has occured while updating genre node and it's relationships: %v", err)
		return err
	}

	return nil
}

func (rs *RecommendationService) SubscriptionBasedRecommendation(ctx context.Context) ([]*entities.SongRecommendation, error) {
	recCtx, recSpan := rs.tr.Start(ctx, "recommendation.sub_based")
	defer recSpan.End()

	userIDStr := middlewares.GetUserIdFromContext(ctx)

	// just to be sure if it's a valid UUID
	_, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrObjectIdCastFailed
	}

	srs, err := rs.r.rr.FindSubscriptionBasedRecommendations(recCtx, userIDStr)
	if err != nil {
		return nil, err
	}

	return srs, nil
}

func (rs *RecommendationService) LikeBasedRecommendation(ctx context.Context) ([]*entities.SongRecommendation, error) {
	recCtx, recSpan := rs.tr.Start(ctx, "recommendation.like_based")
	defer recSpan.End()

	userIDStr := middlewares.GetUserIdFromContext(ctx)

	_, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrObjectIdCastFailed
	}

	lr, err := rs.r.rr.FindLikeBasedRecommendation(recCtx, userIDStr)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (rs *RecommendationService) DeleteSong(e events.SongDeletePayload, ctx context.Context) error {
	recCtx, recSpan := rs.tr.Start(ctx, "recommendation.delete_song")
	defer recSpan.End()

	err := rs.r.rr.DeleteSong(recCtx, e.SongID)
	if err != nil {
		return err
	}

	return nil
}
