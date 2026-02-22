package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
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
	ErrEntityNotFound      = errors.New("song couldn't be found")
	ErrSongNotFound        = errors.New("song couldn't be found")
	ErrInvalidEntityID     = errors.New("error has ocurred while parsing song ID")
	ErrRatingNotFound      = errors.New("rating not found")
	ErrRatingForbidden     = errors.New("rating does not belong to user")
	ErrNoFieldsToUpdate    = errors.New("no fields to update")
	ErrObjectIdCastFailed  = errors.New("failed to convert hex to objectID")
	ErrUpstreamTimeout     = errors.New("upstream service request timed out")
	ErrUpstreamFailure     = errors.New("upstream service returned an internal error")
	ErrUpstreamUnavailable = errors.New("upstream service is temporarily unavailable")
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
	cb  *gobreaker.CircuitBreaker
	tr  trace.Tracer
}

// NewRatingService creates and returns a new RatingService with the provided repository and content entity getter.
func NewRatingService(rr RatingRepository, gcc ContentEntityGetter) *RatingService {
	tr := otel.Tracer("rating-service/rating-service")

	settings := gobreaker.Settings{
		Name: "rating-service",
		// defines the number of request which will be passed through when the circuit breaker is half open
		// on which we are going to decide will we keep the circuit open or close it
		MaxRequests: 3,

		// defines the time window in which the request states will be saved, when the time is up, all request
		// data is being removed
		Interval: 15 * time.Second,

		// amount of time given to the server to get back up, since the content service dependencies aren't slow
		// to start up like cassandra 10 secs is fair
		Timeout: 10 * time.Second,

		// defines the case in which the circuit will be opened, in this case if more than 10 requests have been
		// executed and more than 30% of them failed, we want to open the circuit
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.3
			// return counts.TotalFailures >= 1
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Print("rating service circuit breaker state changed: ", to.String())
		},
	}
	s := RatingService{rr: rr, gcc: gcc, cb: gobreaker.NewCircuitBreaker(settings), tr: tr}

	return &s
}

// CreateRating creates a new rating for a song. It first checks if the song exists by calling the content entity getter.
func (s *RatingService) CreateRating(req *dtos.CreateRatingDto, ctx context.Context) error {
	ratingCtx, ratingSpan := s.tr.Start(ctx, "rating.create")
	defer ratingSpan.End()

	_, err := s.cb.Execute(func() (any, error) {
		entityCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		entityExistenceCtx, entityExistenceSpan := s.tr.Start(entityCtx, "rating.create.song_exists")
		defer entityExistenceSpan.End()

		name, err := s.gcc.GetSong(entityExistenceCtx, req.SongID)
		if err != nil {
			return nil, err
		}
		return name, nil
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			return ErrUpstreamUnavailable
		}

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		//nolint:exhaustive
		switch st.Code() {
		case codes.NotFound:
			return ErrEntityNotFound
		case codes.InvalidArgument:
			return ErrInvalidEntityID
		case codes.DeadlineExceeded:
			return ErrUpstreamTimeout
		default:
			return ErrUpstreamFailure
		}
	}

	userIDstr := middlewares.GetUserIdFromContext(ratingCtx)
	username := middlewares.GetUsernameFromContext(ratingCtx)

	ratingEntity, err := mappers.ToRatingEntity(req.SongID, userIDstr, req.Value, username)
	if err != nil {
		ratingSpan.RecordError(err)
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
	defer findSpan.End()

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
