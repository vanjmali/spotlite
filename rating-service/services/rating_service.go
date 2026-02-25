package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/rating-service/dtos"
	"github.com/vanjmali/spotlite/rating-service/entities"
	"github.com/vanjmali/spotlite/rating-service/mappers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSongNotFound       = errors.New("song couldn't be found")
	ErrInvalidSongID      = errors.New("error has ocurred while parsing song ID")
	ErrUpstreamFailure    = errors.New("error has ocurred while fetching song")
	ErrRatingNotFound     = errors.New("rating not found")
	ErrRatingForbidden    = errors.New("rating does not belong to user")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
	ErrObjectIdCastFailed = errors.New("failed to convert hex to objectID")
)

type RatingRepository interface {
	Create(rating *entities.Rating, ctx context.Context) error
	Delete(ratingID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error)
	FindByID(ctx context.Context, ratingID primitive.ObjectID) (*entities.Rating, error)
	FindRatingsBySongID(ctx context.Context, songID primitive.ObjectID, batchSize int, lastID *primitive.ObjectID) ([]*entities.Rating, string, error)
	FindRatingsByUserID(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Rating, int64, error)
	UpdateByID(ctx context.Context, ratingID primitive.ObjectID, userID primitive.ObjectID, update map[string]any) (*entities.Rating, error)
	GetAverageRatingBySongID(ctx context.Context, songID primitive.ObjectID) (*dtos.SongRatingSummary, error)
}

type ContentEntityGetter interface {
	GetSong(ctx context.Context, songID string) (string, error)
}
type RatingService struct {
	rr  RatingRepository
	gcc ContentEntityGetter
	tr  trace.Tracer
	jsc *events.JetStreamClient
}

// NewRatingService creates and returns a new RatingService with the provided repository and content entity getter.
func NewRatingService(rr RatingRepository, gcc ContentEntityGetter, jsc *events.JetStreamClient) *RatingService {
	tr := otel.Tracer("rating-service/rating-service")
	s := RatingService{rr: rr, gcc: gcc, tr: tr, jsc: jsc}

	return &s
}

// CreateRating creates a new rating for a song. It first checks if the song exists by calling the content entity getter.
func (s *RatingService) CreateRating(req *dtos.CreateRatingDto, ctx context.Context) error {
	ratingCtx, ratingSpan := s.tr.Start(ctx, "rating.create")
	defer ratingSpan.End()

	songExistsCtx, songExistsSpan := s.tr.Start(ratingCtx, "rating.create.exists")
	defer songExistsSpan.End()

	_, err := s.gcc.GetSong(songExistsCtx, req.SongID)
	if err != nil {
		songExistsSpan.RecordError(err)

		st, ok := status.FromError(err)
		if !ok {
			return err
		}
		// TODO: Handle different types of errors with resiliency mechanisms
		//nolint:exhaustive
		switch st.Code() {
		case codes.NotFound:
			return ErrSongNotFound
		case codes.InvalidArgument:
			return ErrInvalidSongID
		default:
			return ErrUpstreamFailure
		}
	}

	userIDstr := middlewares.GetUserIdFromContext(songExistsCtx)
	username := middlewares.GetUsernameFromContext(songExistsCtx)

	ratingEntity, err := mappers.ToRatingEntity(req.SongID, userIDstr, req.Value, username)
	if err != nil {
		songExistsSpan.RecordError(err)
		return err
	}

	createCtx, createSpan := s.tr.Start(ratingCtx, "rating.create.repo")
	defer createSpan.End()
	err = s.rr.Create(ratingEntity, createCtx)
	if err != nil {
		createSpan.RecordError(err)
		return err
	}

	return nil
}

// DeleteRating deletes a rating by its ID. It first checks if the rating exists and belongs to the user making the request before deleting it.
func (s *RatingService) DeleteRating(ratingID primitive.ObjectID, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "rating.delete")
	defer span.End()

	userIDStr := middlewares.GetUserIdFromContext(ctx)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	existing, err := s.rr.FindByID(ctx, ratingID)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrRatingNotFound
		}
		return err
	}

	if existing.UserID != userID {
		span.RecordError(ErrRatingForbidden)
		return ErrRatingForbidden
	}

	deleteCtx, deleteSpan := s.tr.Start(ctx, "rating.delete.repo")
	defer deleteSpan.End()

	deletedCount, err := s.rr.Delete(ratingID, userID, deleteCtx)
	if err != nil {
		deleteSpan.RecordError(err)
		return err
	}

	if deletedCount != 1 {
		return ErrRatingNotFound
	}

	return nil
}

// GetRatingBySong retrieves ratings for a specific song with pagination support. It accepts the song ID, batch size, and an optional cursor for pagination.
func (s *RatingService) GetRatingBySong(ctx context.Context, songIDStr string, batchSize int, cursor string) ([]*entities.Rating, string, error) {
	getCtx, getSpan := s.tr.Start(ctx, "rating.get_by_song")
	defer getSpan.End()

	findCtx, findSpan := s.tr.Start(getCtx, "rating.get_by_song.find")
	defer findSpan.End()

	songID, err := primitive.ObjectIDFromHex(songIDStr)
	if err != nil {
		return nil, "", ErrObjectIdCastFailed
	}

	var lastID *primitive.ObjectID
	if cursor != "" {
		objID, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			return nil, "", ErrObjectIdCastFailed
		}
		lastID = &objID
	}

	ratings, nextCursor, err := s.rr.FindRatingsBySongID(findCtx, songID, batchSize, lastID)
	if err != nil {
		return nil, "", err
	}

	return ratings, nextCursor, nil
}

// RatingsQuery represents the query parameters for retrieving ratings by user.
type RatingsQuery struct {
	Page   int
	Size   int
	UserID string
}

// GetRatingByUser retrieves ratings made by a specific user with pagination support.
func (s *RatingService) GetRatingByUser(ctx context.Context, q RatingsQuery) (*dtos.RatingListResponseDto, error) {
	getCtx, getSpan := s.tr.Start(ctx, "rating.get_by_user")
	defer getSpan.End()

	filter := bson.M{}

	if q.UserID != "" {
		userID, err := primitive.ObjectIDFromHex(q.UserID)
		if err != nil {
			return nil, ErrObjectIdCastFailed
		}
		filter["user_id"] = userID
	}

	p := pagination.NewPagination(q.Page, q.Size)
	items, total, err := s.rr.FindRatingsByUserID(getCtx, filter, p.Skip(), p.Limit())
	if err != nil {
		getSpan.RecordError(err)
		return nil, err
	}

	return &dtos.RatingListResponseDto{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}

// UpdateRating updates an existing rating. It first checks if the rating exists and belongs to the user making the request before applying the updates.
func (s *RatingService) UpdateRating(ctx context.Context, ratingIdStr string, dto dtos.UpdateRatingDto) (*entities.Rating, error) {
	updateCtx, updateSpan := s.tr.Start(ctx, "rating.update")
	defer updateSpan.End()

	ratingID, err := primitive.ObjectIDFromHex(ratingIdStr)
	if err != nil {
		updateSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	userIDStr := middlewares.GetUserIdFromContext(updateCtx)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		updateSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	update := make(map[string]any)
	if dto.Value != nil {
		update["value"] = *dto.Value
		update["is_edited"] = true
	}
	if len(update) == 0 {
		updateSpan.RecordError(ErrNoFieldsToUpdate)
		return nil, ErrNoFieldsToUpdate
	}

	findCtx, findSpan := s.tr.Start(ctx, "rating.update.find")
	defer findSpan.End()

	existing, err := s.rr.FindByID(findCtx, ratingID)
	if err != nil {
		findSpan.RecordError(err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRatingNotFound
		}
		return nil, err
	}

	if existing.UserID != userID {
		updateSpan.RecordError(ErrRatingForbidden)
		return nil, ErrRatingForbidden
	}

	repoCtx, repoSpan := s.tr.Start(ctx, "rating.update.repo")
	defer repoSpan.End()

	rating, err := s.rr.UpdateByID(repoCtx, ratingID, userID, update)
	if err != nil {
		repoSpan.RecordError(err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRatingNotFound
		}
		return nil, err
	}

	return rating, nil
}

// GetAverageRatingBySongID retrieves the average rating for a specific song.
func (s *RatingService) GetAverageRatingBySongID(ctx context.Context, songIDStr string) (*dtos.SongRatingSummary, error) {
	avgCtx, avgSpan := s.tr.Start(ctx, "rating.average")
	defer avgSpan.End()

	songID, err := primitive.ObjectIDFromHex(songIDStr)
	if err != nil {
		avgSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	summary, err := s.rr.GetAverageRatingBySongID(avgCtx, songID)
	if err != nil {
		avgSpan.RecordError(err)
		return nil, err
	}

	return summary, nil
}
