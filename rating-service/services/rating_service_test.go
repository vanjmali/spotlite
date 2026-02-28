package services

import (
	"context"
	"errors"
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/rating-service/dtos"
	"github.com/vanjmali/spotlite/rating-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errFakeRepoSentinel = errors.New("fake repo sentinel")

type fakeRatingRepo struct {
	createCalled bool
}

func (f *fakeRatingRepo) Create(rating *entities.Rating, ctx context.Context) error {
	f.createCalled = true
	return nil
}

func (f *fakeRatingRepo) Delete(ratingID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeRatingRepo) FindByID(ctx context.Context, ratingID primitive.ObjectID) (*entities.Rating, error) {
	return nil, errFakeRepoSentinel
}

func (f *fakeRatingRepo) FindRatingsBySongID(
	ctx context.Context,
	songID primitive.ObjectID,
	batchSize int,
	lastID *primitive.ObjectID,
) ([]*entities.Rating, string, error) {
	return nil, "", nil
}

func (f *fakeRatingRepo) FindRatingsByUserID(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Rating, int64, error) {
	return nil, 0, nil
}

func (f *fakeRatingRepo) UpdateByID(
	ctx context.Context,
	ratingID primitive.ObjectID,
	userID primitive.ObjectID,
	update map[string]any,
) (*entities.Rating, error) {
	return nil, errFakeRepoSentinel
}

func (f *fakeRatingRepo) GetAverageRatingBySongID(ctx context.Context, songID primitive.ObjectID) (*dtos.SongRatingSummary, error) {
	return nil, errFakeRepoSentinel
}

type fakeSongGetter struct {
	getSongFn func(ctx context.Context, songID string) (string, error)
}

func (f *fakeSongGetter) GetSong(ctx context.Context, songID string) (string, error) {
	if f.getSongFn != nil {
		return f.getSongFn(ctx, songID)
	}
	return "name", nil
}

func TestCreateRatingMapsTimeout(t *testing.T) {
	repo := &fakeRatingRepo{}
	getter := &fakeSongGetter{
		getSongFn: func(context.Context, string) (string, error) {
			return "", status.Error(codes.DeadlineExceeded, "timeout")
		},
	}
	jsc := &events.JetStreamClient{}
	svc := NewRatingService(repo, getter, jsc)

	err := svc.CreateRating(&dtos.CreateRatingDto{SongID: primitive.NewObjectID().Hex(), Value: 5}, context.Background())

	require.ErrorIs(t, err, ErrUpstreamTimeout)
	require.False(t, repo.createCalled)
}

func TestCreateRatingMapsUnavailable(t *testing.T) {
	repo := &fakeRatingRepo{}
	getter := &fakeSongGetter{
		getSongFn: func(context.Context, string) (string, error) {
			return "", gobreaker.ErrOpenState
		},
	}
	jsc := &events.JetStreamClient{}
	svc := NewRatingService(repo, getter, jsc)

	err := svc.CreateRating(&dtos.CreateRatingDto{SongID: primitive.NewObjectID().Hex(), Value: 5}, context.Background())

	require.ErrorIs(t, err, ErrUpstreamUnavailable)
	require.False(t, repo.createCalled)
}

func TestCreateRatingMapsThrottled(t *testing.T) {
	repo := &fakeRatingRepo{}
	getter := &fakeSongGetter{
		getSongFn: func(context.Context, string) (string, error) {
			return "", gobreaker.ErrTooManyRequests
		},
	}
	jsc := &events.JetStreamClient{}
	svc := NewRatingService(repo, getter, jsc)

	err := svc.CreateRating(&dtos.CreateRatingDto{SongID: primitive.NewObjectID().Hex(), Value: 5}, context.Background())

	require.ErrorIs(t, err, ErrUpstreamThrottled)
	require.False(t, repo.createCalled)
}
