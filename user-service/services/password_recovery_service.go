package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/vanjmali/spotlite/common-lib/clock"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRecoveryToken = errors.New("invalid or expired recovery token")
	ErrRecoveryTokenExpired = errors.New("recovery token expired")
	ErrRecoveryTokenUsed    = errors.New("recovery token already used")
)

// PasswordRecoveryRepository defines persistence methods for password recovery tokens.
type PasswordRecoveryRepository interface {
	SaveRecoveryToken(ctx context.Context, token *entities.PasswordRecoveryToken) error
	FindValidToken(ctx context.Context, tokenSearchable string) (*entities.PasswordRecoveryToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID primitive.ObjectID) error
	InvalidateAllTokensForUser(ctx context.Context, userID primitive.ObjectID) error
}

// PasswordRecoveryService handles password recovery operations including token generation,
// validation, and password reset.
type PasswordRecoveryService struct {
	r  UserRepository
	pr PasswordRecoveryRepository
	ms MailSender
	c  clock.Clock
	tr trace.Tracer
}

// NewPasswordRecoveryService builds a PasswordRecoveryService with required dependencies.
func NewPasswordRecoveryService(r UserRepository, pr PasswordRecoveryRepository, ms MailSender) *PasswordRecoveryService {
	tr := otel.Tracer("user-service/password-recovery")
	s := PasswordRecoveryService{r: r, pr: pr, ms: ms, c: clock.RealClock{}, tr: tr}
	return &s
}

// RequestPasswordReset generates a recovery token and sends a reset email.
func (s *PasswordRecoveryService) RequestPasswordReset(ctx context.Context, email string) error {
	ctx, span := s.tr.Start(ctx, "password_recovery.request_reset")
	defer span.End()

	// 1. Check if user exists
	user, err := s.r.FindUserByEmail(ctx, email)
	if err != nil {
		// Don't leak if email exists - return generic success
		logging.Auditf(ctx, "password recovery request for non-existent email")
		return nil
	}

	// 2. Invalidate any existing recovery tokens for this user
	if err := s.pr.InvalidateAllTokensForUser(ctx, user.ID); err != nil {
		logging.Warnf(ctx, "failed to invalidate previous recovery tokens: %v", err)
		// Continue anyway, not a critical failure
	}

	// 3. Generate secure random token
	plainToken, err := s.generateSecureToken()
	if err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to generate recovery token: %v", err)
		return err
	}

	// 4. Create two hashes:
	// - bcrypt for security (not searchable)
	// - SHA256 for database lookup
	tokenHashBcrypt, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to hash recovery token: %v", err)
		return err
	}

	tokenSearchable := s.hashTokenSHA256(plainToken)

	// 5. Create recovery record (expires in 15 minutes)
	recoveryToken := &entities.PasswordRecoveryToken{
		UserID:          user.ID,
		Email:           email,
		TokenHash:       string(tokenHashBcrypt),
		TokenSearchable: tokenSearchable,
		ExpiresAt:       s.c.Now().Add(15 * time.Minute),
		CreatedAt:       s.c.Now(),
	}

	if err := s.pr.SaveRecoveryToken(ctx, recoveryToken); err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to save recovery token: %v", err)
		return err
	}

	// 6. Send email with magic link (send plaintext token, not hashed)
	if err := s.ms.SendPasswordResetEmail(email, plainToken); err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to send password reset email: %v", err)
		return err
	}

	return nil
}

// ValidateRecoveryToken checks if the token is valid, unused, and not expired.
func (s *PasswordRecoveryService) ValidateRecoveryToken(ctx context.Context, token string) (*entities.PasswordRecoveryToken, error) {
	ctx, span := s.tr.Start(ctx, "password_recovery.validate_token")
	defer span.End()

	// Hash the provided token using SHA256 for lookup
	tokenSearchable := s.hashTokenSHA256(token)

	recoveryToken, err := s.pr.FindValidToken(ctx, tokenSearchable)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInvalidRecoveryToken
	}

	// Check if token is expired
	if s.c.Now().After(recoveryToken.ExpiresAt) {
		return nil, ErrRecoveryTokenExpired
	}

	// Check if token already used
	if recoveryToken.UsedAt != nil {
		return nil, ErrRecoveryTokenUsed
	}

	// Verify bcrypt hash matches (additional security)
	if err := bcrypt.CompareHashAndPassword([]byte(recoveryToken.TokenHash), []byte(token)); err != nil {
		span.RecordError(err)
		logging.Warnf(ctx, "recovery token hash mismatch: %v", err)
		return nil, ErrInvalidRecoveryToken
	}

	return recoveryToken, nil
}

// ResetPassword validates the token and updates the user's password.
func (s *PasswordRecoveryService) ResetPassword(ctx context.Context, reqDto *dtos.ResetPasswordDto) error {
	ctx, span := s.tr.Start(ctx, "password_recovery.reset_password")
	defer span.End()

	// 1. Validate and retrieve token
	recoveryToken, err := s.ValidateRecoveryToken(ctx, reqDto.Token)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// 2. Find user
	user, err := s.r.FindUserByEmail(ctx, recoveryToken.Email)
	if err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "user not found for recovery: %v", err)
		return ErrUserNotFound
	}

	// 3. Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(reqDto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to hash new password: %v", err)
		return err
	}

	// 4. Update user password
	now := s.c.Now()
	expiresAt := now.AddDate(0, 3, 0) // Password expires in 3 months
	if err := s.r.SetHashPassowrd(ctx, user.ID, string(passwordHash), now, expiresAt); err != nil {
		span.RecordError(err)
		logging.Errorf(ctx, "failed to update user password: %v", err)
		return err
	}

	// 5. Mark token as used
	if err := s.pr.MarkTokenAsUsed(ctx, recoveryToken.ID); err != nil {
		logging.Warnf(ctx, "failed to mark recovery token as used: %v", err)
		// Not critical, continue
	}

	return nil
}

// generateSecureToken creates a cryptographically secure random token.
func (s *PasswordRecoveryService) generateSecureToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// hashTokenSHA256 creates a SHA256 hash of the token for database lookup.
func (s *PasswordRecoveryService) hashTokenSHA256(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
