package services

import (
	"context"
	"errors"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrRefreshInvalid = errors.New("invalid refresh token")
)

type RefreshTokenService struct {
	r *repositories.RefreshTokenRepository
}

func NewRefreshTokenService(r repositories.RefreshTokenRepository) *RefreshTokenService {
	s := RefreshTokenService{r: &r}

	return &s
}

func (s *RefreshTokenService) IssueRefreshToken(ctx context.Context, userID primitive.ObjectID) (string, error) {
	raw, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	doc := entities.RefreshToken{
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(raw),
		RevokedAt: nil,
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	_, err = s.r.InsertRefreshToken(ctx, doc)
	if err != nil {
		return "", err
	}

	return raw, err
}

func (s *RefreshTokenService) GetRefreshTokenId(ctx context.Context, refreshToken string) (primitive.ObjectID, error) {
	hash := auth.HashRefreshToken(refreshToken)

	old, err := s.r.FindActiveRefreshByHash(ctx, hash)
	if err != nil || old == nil {
		return primitive.NilObjectID, ErrRefreshInvalid
	}

	if time.Now().After(old.ExpiresAt) {
		return primitive.NilObjectID, ErrRefreshInvalid
	}

	return old.UserID, nil
}
