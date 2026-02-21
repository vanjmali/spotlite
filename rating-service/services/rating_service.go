package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/rating-service/dtos"
	"github.com/vanjmali/spotlite/rating-service/entities"
	"github.com/vanjmali/spotlite/rating-service/mappers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSongNotFound    = errors.New("song couldn't be found")
	ErrInvalidSongID   = errors.New("error has ocurred while parsing song id")
	ErrUpstreamFailure = errors.New("error has ocurred while fetching song")
	ErrRatingNotFound  = errors.New("rating not found")
)

type RatingRepository interface {
	Create(rating *entities.Rating, ctx context.Context) error
	Delete(ratingID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error)
	FindRatingsBySongID(ctx context.Context, songID primitive.ObjectID, batchSize int, lastID *primitive.ObjectID) ([]*entities.Rating, string, error)
}

type ContentEntityGetter interface {
	GetSong(ctx context.Context, songID string) (string, error)
}
type RatingService struct {
	rr  RatingRepository
	gcc ContentEntityGetter
	tr  trace.Tracer
}

func NewRatingService(rr RatingRepository, gcc ContentEntityGetter) *RatingService {
	tr := otel.Tracer("rating-service/rating-service")
	s := RatingService{rr: rr, gcc: gcc, tr: tr}

	return &s
}

func (s *RatingService) CreateRating(req *dtos.CreateRatingDto, ctx context.Context) error {
	ratingCtx, ratingSpan := s.tr.Start(ctx, "rating.create_rating")
	defer ratingSpan.End()

	songExistsCtx, songExistsSpan := s.tr.Start(ratingCtx, "rating.create_rating.song_exists")
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

	createCtx, createSpan := s.tr.Start(ratingCtx, "rating.create_rating.create")
	defer createSpan.End()
	err = s.rr.Create(ratingEntity, createCtx)
	if err != nil {
		createSpan.RecordError(err)
		return err
	}

	return nil
}

func (s *RatingService) DeleteRating(ratingID primitive.ObjectID, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "rating.delete_rating")
	defer span.End()

	userIDStr := middlewares.GetUserIdFromContext(ctx)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		span.RecordError(err)
		return err
	}

	deleteCtx, deleteSpan := s.tr.Start(ctx, "rating.delete_rating.delete")
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

func (s *RatingService) GetRatingBySong(ctx context.Context, songIDStr string, batchSize int, cursor string) ([]*entities.Rating, string, error) {
	getCtx, getSpan := s.tr.Start(ctx, "rating.get_by_song")
	defer getSpan.End()

	findCtx, findSpan := s.tr.Start(getCtx, "rating.get_by_song.find")
	defer findSpan.End()

	songID, err := primitive.ObjectIDFromHex(songIDStr)
	if err != nil {
		return nil, "", ErrInvalidSongID
	}

	var lastID *primitive.ObjectID
	if cursor != "" {
		objID, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			return nil, "", ErrInvalidSongID
		}
		lastID = &objID
	}

	ratings, nextCursor, err := s.rr.FindRatingsBySongID(findCtx, songID, batchSize, lastID)
	if err != nil {
		return nil, "", err
	}

	return ratings, nextCursor, nil
}
