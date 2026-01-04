package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeRefreshTokenRepo struct {
	insertFn       func(context.Context, entities.RefreshToken) (primitive.ObjectID, error)
	findActiveByFn func(context.Context, string) (*entities.RefreshToken, error)
	revokeFn       func(context.Context, primitive.ObjectID, time.Time, primitive.ObjectID) error
	insertCalled   bool
	insertedToken  entities.RefreshToken
	findActiveHash string
}

func (f *fakeRefreshTokenRepo) InsertToken(ctx context.Context, rt entities.RefreshToken) (primitive.ObjectID, error) {
	f.insertCalled = true
	f.insertedToken = rt
	if f.insertFn != nil {
		return f.insertFn(ctx, rt)
	}
	return primitive.NewObjectID(), nil
}

func (f *fakeRefreshTokenRepo) FindActiveByHash(ctx context.Context, hash string) (*entities.RefreshToken, error) {
	f.findActiveHash = hash
	if f.findActiveByFn != nil {
		return f.findActiveByFn(ctx, hash)
	}
	return &entities.RefreshToken{}, nil
}

func (f *fakeRefreshTokenRepo) RevokeByID(
	ctx context.Context,
	id primitive.ObjectID,
	when time.Time,
	replacedBy primitive.ObjectID,
) error {
	if f.revokeFn != nil {
		return f.revokeFn(ctx, id, when, replacedBy)
	}
	return nil
}

func TestRefreshTokenServiceIssueRefreshToken(t *testing.T) {
	repo := &fakeRefreshTokenRepo{}
	svc := NewRefreshTokenService(repo)
	userID := primitive.NewObjectID()

	start := time.Now()
	raw, _, err := svc.IssueRefreshToken(context.Background(), userID)

	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.True(t, repo.insertCalled)
	require.Equal(t, userID, repo.insertedToken.UserID)
	require.Nil(t, repo.insertedToken.RevokedAt)
	require.Equal(t, auth.HashRefreshToken(raw), repo.insertedToken.TokenHash)
	require.True(t, repo.insertedToken.ExpiresAt.After(start.Add(29*24*time.Hour)))
	require.True(t, repo.insertedToken.ExpiresAt.Before(start.Add(31*24*time.Hour)))
}

func TestRefreshTokenServiceIssueRefreshTokenInsertFails(t *testing.T) {
	repo := &fakeRefreshTokenRepo{
		insertFn: func(context.Context, entities.RefreshToken) (primitive.ObjectID, error) {
			return primitive.NilObjectID, errors.New("insert failed")
		},
	}
	svc := NewRefreshTokenService(repo)

	raw, _, err := svc.IssueRefreshToken(context.Background(), primitive.NewObjectID())

	require.Error(t, err)
	require.Empty(t, raw)
}

func TestRefreshTokenServiceGetRefreshTokenIdInvalid(t *testing.T) {
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return &entities.RefreshToken{}, nil
		},
	}
	svc := NewRefreshTokenService(repo)

	userID, err := svc.GetRefreshTokenId(context.Background(), "token")

	require.ErrorIs(t, err, ErrRefreshInvalid)
	require.Equal(t, primitive.NilObjectID, userID)
}

func TestRefreshTokenServiceGetRefreshTokenIdExpired(t *testing.T) {
	oldUserID := primitive.NewObjectID()
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return &entities.RefreshToken{
				UserID:    oldUserID,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			}, nil
		},
	}
	svc := NewRefreshTokenService(repo)

	userID, err := svc.GetRefreshTokenId(context.Background(), "token")

	require.ErrorIs(t, err, ErrRefreshInvalid)
	require.Equal(t, primitive.NilObjectID, userID)
}

func TestRefreshTokenServiceGetRefreshTokenIdSuccess(t *testing.T) {
	oldUserID := primitive.NewObjectID()
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return &entities.RefreshToken{
				UserID:    oldUserID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}
	svc := NewRefreshTokenService(repo)

	userID, err := svc.GetRefreshTokenId(context.Background(), "token")

	require.NoError(t, err)
	require.Equal(t, oldUserID, userID)
}

func TestRefreshTokenServiceGetRefreshTokenIdRepoError(t *testing.T) {
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewRefreshTokenService(repo)

	userID, err := svc.GetRefreshTokenId(context.Background(), "token")

	require.ErrorIs(t, err, ErrRefreshInvalid)
	require.Equal(t, primitive.NilObjectID, userID)
}

func TestRefreshTokenServiceRevokeRefreshTokenSuccess(t *testing.T) {
	tokenID := primitive.NewObjectID()
	revokeCalled := false
	var revokedID primitive.ObjectID
	var revokedAt time.Time
	var replacedBy primitive.ObjectID

	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return &entities.RefreshToken{
				ID:        tokenID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
		revokeFn: func(_ context.Context, id primitive.ObjectID, when time.Time, replaced primitive.ObjectID) error {
			revokeCalled = true
			revokedID = id
			revokedAt = when
			replacedBy = replaced
			return nil
		},
	}
	svc := NewRefreshTokenService(repo)

	err := svc.RevokeRefreshToken(context.Background(), "token")

	require.NoError(t, err)
	require.True(t, revokeCalled)
	require.Equal(t, tokenID, revokedID)
	require.False(t, revokedAt.IsZero())
	require.Equal(t, primitive.NilObjectID, replacedBy)
}

func TestRefreshTokenServiceRevokeRefreshTokenInvalid(t *testing.T) {
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewRefreshTokenService(repo)

	err := svc.RevokeRefreshToken(context.Background(), "token")

	require.ErrorIs(t, err, ErrRefreshInvalid)
}

func TestRefreshTokenServiceRevokeRefreshTokenRepoError(t *testing.T) {
	repo := &fakeRefreshTokenRepo{
		findActiveByFn: func(context.Context, string) (*entities.RefreshToken, error) {
			return &entities.RefreshToken{
				ID:        primitive.NewObjectID(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
		revokeFn: func(context.Context, primitive.ObjectID, time.Time, primitive.ObjectID) error {
			return errors.New("revoke failed")
		},
	}
	svc := NewRefreshTokenService(repo)

	err := svc.RevokeRefreshToken(context.Background(), "token")

	require.Error(t, err)
}
