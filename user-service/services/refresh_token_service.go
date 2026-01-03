package services

import (
	"context"
	"errors"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrRefreshInvalid = errors.New("invalid refresh token")

type RefreshTokenService struct {
	r *repositories.RefreshTokenRepository

	tr trace.Tracer
}

func NewRefreshTokenService(r repositories.RefreshTokenRepository) *RefreshTokenService {
	tr := otel.Tracer("user-service/refresh-token-service")
	s := RefreshTokenService{r: &r, tr: tr}

	return &s
}

func (s *RefreshTokenService) IssueRefreshToken(ctx context.Context, userID primitive.ObjectID) (string, time.Time, error) {
	ctx, span := s.tr.Start(ctx, "refresh_token.issue")
	defer span.End()

	raw, err := auth.GenerateRefreshToken()
	if err != nil {
		span.RecordError(err)
		return "", time.Time{}, err
	}
	now := time.Now()
	doc := entities.RefreshToken{
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(raw),
		RevokedAt: nil,
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	_, err = s.r.InsertToken(ctx, doc)
	if err != nil {
		span.RecordError(err)
		return "", time.Time{}, err
	}

	return raw, doc.ExpiresAt, err
}

func (s *RefreshTokenService) GetRefreshTokenId(ctx context.Context, refreshToken string) (primitive.ObjectID, error) {
	ctx, span := s.tr.Start(ctx, "refresh_token.verify")
	defer span.End()

	hash := auth.HashRefreshToken(refreshToken)

	old, err := s.r.FindActiveByHash(ctx, hash)
	if err != nil || old == nil {
		if err != nil {
			span.RecordError(err)
		}
		return primitive.NilObjectID, ErrRefreshInvalid
	}

	if time.Now().After(old.ExpiresAt) {
		return primitive.NilObjectID, ErrRefreshInvalid
	}

	return old.UserID, nil
}

func (s *RefreshTokenService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	ctx, span := s.tr.Start(ctx, "refresh_token.revoke")
	defer span.End()

	hash := auth.HashRefreshToken(refreshToken)
	old, err := s.r.FindActiveByHash(ctx, hash)
	if err != nil || old == nil {
		if err != nil {
			span.RecordError(err)
		}
		return ErrRefreshInvalid
	}

	now := time.Now()
	if err := s.r.RevokeByID(ctx, old.ID, now, primitive.NilObjectID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
