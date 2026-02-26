package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type fakePasswordRecoveryRepo struct {
	saveRecoveryTokenFn       func(context.Context, *entities.PasswordRecoveryToken) error
	findValidTokenFn          func(context.Context, string) (*entities.PasswordRecoveryToken, error)
	markTokenAsUsedFn         func(context.Context, primitive.ObjectID) error
	invalidateAllTokensForFn  func(context.Context, primitive.ObjectID) error
	saveRecoveryTokenCalled   bool
	findValidTokenCalled      bool
	markTokenAsUsedCalled     bool
	invalidateAllTokensCalled bool
}

func (f *fakePasswordRecoveryRepo) SaveRecoveryToken(ctx context.Context, token *entities.PasswordRecoveryToken) error {
	f.saveRecoveryTokenCalled = true
	if f.saveRecoveryTokenFn != nil {
		return f.saveRecoveryTokenFn(ctx, token)
	}
	return nil
}

func (f *fakePasswordRecoveryRepo) FindValidToken(ctx context.Context, tokenSearchable string) (*entities.PasswordRecoveryToken, error) {
	f.findValidTokenCalled = true
	if f.findValidTokenFn != nil {
		return f.findValidTokenFn(ctx, tokenSearchable)
	}
	return &entities.PasswordRecoveryToken{}, nil
}

func (f *fakePasswordRecoveryRepo) MarkTokenAsUsed(ctx context.Context, tokenID primitive.ObjectID) error {
	f.markTokenAsUsedCalled = true
	if f.markTokenAsUsedFn != nil {
		return f.markTokenAsUsedFn(ctx, tokenID)
	}
	return nil
}

func (f *fakePasswordRecoveryRepo) InvalidateAllTokensForUser(ctx context.Context, userID primitive.ObjectID) error {
	f.invalidateAllTokensCalled = true
	if f.invalidateAllTokensForFn != nil {
		return f.invalidateAllTokensForFn(ctx, userID)
	}
	return nil
}

type fakeUserRepositoryForPasswordRecovery struct {
	findUserByEmailFn func(context.Context, string) (*entities.User, error)
	setHashPasswordFn func(context.Context, primitive.ObjectID, string, time.Time, time.Time) error
}

func (f *fakeUserRepositoryForPasswordRecovery) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	return &entities.User{}, nil
}

func (f *fakeUserRepositoryForPasswordRecovery) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	if f.findUserByEmailFn != nil {
		return f.findUserByEmailFn(ctx, email)
	}
	return &entities.User{ID: primitive.NewObjectID(), Email: email}, nil
}

func (f *fakeUserRepositoryForPasswordRecovery) SetHashPassowrd(
	ctx context.Context,
	userID primitive.ObjectID,
	passwordHash string,
	changedAt time.Time,
	expiresAt time.Time,
) error {
	if f.setHashPasswordFn != nil {
		return f.setHashPasswordFn(ctx, userID, passwordHash, changedAt, expiresAt)
	}
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (f *fakeUserRepositoryForPasswordRecovery) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (f *fakeUserRepositoryForPasswordRecovery) Create(ctx context.Context, user entities.User) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) SetLoginOtp(ctx context.Context, userId primitive.ObjectID, hash string, expiry time.Time) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) FindUsersForExpiryNotification(
	ctx context.Context,
	daysUntilExpiry int,
	batchSize int,
	lastID string,
) ([]*entities.User, string, error) {
	return nil, "", nil
}

func (f *fakeUserRepositoryForPasswordRecovery) UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	return &entities.User{}, nil
}

func (f *fakeUserRepositoryForPasswordRecovery) ActiveAndRevokeToken(ctx context.Context, token string) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) UpdateVerificationToken(
	ctx context.Context,
	userID primitive.ObjectID,
	token string,
) error {
	return nil
}

func (f *fakeUserRepositoryForPasswordRecovery) Delete(ctx context.Context, userID primitive.ObjectID) error {
	return nil
}

func TestPasswordRecoveryServiceRequestPasswordResetUserNotFound(t *testing.T) {
	userRepo := &fakeUserRepositoryForPasswordRecovery{
		findUserByEmailFn: func(context.Context, string) (*entities.User, error) {
			return nil, errors.New("user not found")
		},
	}
	recoveryRepo := &fakePasswordRecoveryRepo{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	err := svc.RequestPasswordReset(context.Background(), "nonexistent@example.com")

	require.NoError(t, err)
	require.False(t, recoveryRepo.saveRecoveryTokenCalled)
	require.False(t, mailService.sendPasswordResetCalled)
}

func TestPasswordRecoveryServiceRequestPasswordResetSuccess(t *testing.T) {
	userID := primitive.NewObjectID()
	userRepo := &fakeUserRepositoryForPasswordRecovery{
		findUserByEmailFn: func(context.Context, string) (*entities.User, error) {
			return &entities.User{ID: userID, Email: "user@example.com"}, nil
		},
	}
	recoveryRepo := &fakePasswordRecoveryRepo{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	err := svc.RequestPasswordReset(context.Background(), "user@example.com")

	require.NoError(t, err)
	require.True(t, recoveryRepo.invalidateAllTokensCalled)
	require.True(t, recoveryRepo.saveRecoveryTokenCalled)
	require.True(t, mailService.sendPasswordResetCalled)
}

func TestPasswordRecoveryServiceRequestPasswordResetSaveTokenFails(t *testing.T) {
	userID := primitive.NewObjectID()
	userRepo := &fakeUserRepositoryForPasswordRecovery{
		findUserByEmailFn: func(context.Context, string) (*entities.User, error) {
			return &entities.User{ID: userID, Email: "user@example.com"}, nil
		},
	}
	recoveryRepo := &fakePasswordRecoveryRepo{
		saveRecoveryTokenFn: func(context.Context, *entities.PasswordRecoveryToken) error {
			return errors.New("save failed")
		},
	}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	err := svc.RequestPasswordReset(context.Background(), "user@example.com")

	require.Error(t, err)
	require.False(t, mailService.sendPasswordResetCalled)
}

func TestPasswordRecoveryServiceValidateRecoveryTokenInvalid(t *testing.T) {
	recoveryRepo := &fakePasswordRecoveryRepo{
		findValidTokenFn: func(context.Context, string) (*entities.PasswordRecoveryToken, error) {
			return nil, errors.New("token not found")
		},
	}
	userRepo := &fakeUserRepositoryForPasswordRecovery{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	_, err := svc.ValidateRecoveryToken(context.Background(), "invalid-token")

	require.ErrorIs(t, err, ErrInvalidRecoveryToken)
}

func TestPasswordRecoveryServiceValidateRecoveryTokenExpired(t *testing.T) {
	recoveryRepo := &fakePasswordRecoveryRepo{
		findValidTokenFn: func(context.Context, string) (*entities.PasswordRecoveryToken, error) {
			return &entities.PasswordRecoveryToken{
				ID:        primitive.NewObjectID(),
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			}, nil
		},
	}
	userRepo := &fakeUserRepositoryForPasswordRecovery{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	_, err := svc.ValidateRecoveryToken(context.Background(), "expired-token")

	require.ErrorIs(t, err, ErrRecoveryTokenExpired)
}

func TestPasswordRecoveryServiceValidateRecoveryTokenAlreadyUsed(t *testing.T) {
	usedAt := time.Now()
	recoveryRepo := &fakePasswordRecoveryRepo{
		findValidTokenFn: func(context.Context, string) (*entities.PasswordRecoveryToken, error) {
			return &entities.PasswordRecoveryToken{
				ID:        primitive.NewObjectID(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UsedAt:    &usedAt,
			}, nil
		},
	}
	userRepo := &fakeUserRepositoryForPasswordRecovery{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	_, err := svc.ValidateRecoveryToken(context.Background(), "used-token")

	require.ErrorIs(t, err, ErrRecoveryTokenUsed)
}

func TestPasswordRecoveryServiceResetPasswordSuccess(t *testing.T) {
	userID := primitive.NewObjectID()
	tokenID := primitive.NewObjectID()
	plainToken := "valid-token-12345"
	hashedToken, _ := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)

	recoveryRepo := &fakePasswordRecoveryRepo{
		findValidTokenFn: func(context.Context, string) (*entities.PasswordRecoveryToken, error) {
			return &entities.PasswordRecoveryToken{
				ID:        tokenID,
				UserID:    userID,
				Email:     "user@example.com",
				TokenHash: string(hashedToken),
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UsedAt:    nil,
			}, nil
		},
	}
	userRepo := &fakeUserRepositoryForPasswordRecovery{
		findUserByEmailFn: func(context.Context, string) (*entities.User, error) {
			return &entities.User{ID: userID, Email: "user@example.com"}, nil
		},
	}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	err := svc.ResetPassword(context.Background(), &dtos.ResetPasswordDto{
		Token:       plainToken,
		NewPassword: "NewSecurePassword123!",
	})

	require.NoError(t, err)
	require.True(t, recoveryRepo.markTokenAsUsedCalled)
}

func TestPasswordRecoveryServiceResetPasswordInvalidToken(t *testing.T) {
	recoveryRepo := &fakePasswordRecoveryRepo{
		findValidTokenFn: func(context.Context, string) (*entities.PasswordRecoveryToken, error) {
			return nil, errors.New("token not found")
		},
	}
	userRepo := &fakeUserRepositoryForPasswordRecovery{}
	mailService := &fakeMailService{}
	svc := NewPasswordRecoveryService(userRepo, recoveryRepo, mailService)

	err := svc.ResetPassword(context.Background(), &dtos.ResetPasswordDto{
		Token:       "invalid-token",
		NewPassword: "NewSecurePassword123!",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidRecoveryToken)
}
