package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/clock"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken             = errors.New("username is already taken")
	ErrEmailTaken                = errors.New("email is already taken")
	ErrExpiredPassword           = errors.New("your password is expired")
	ErrUserInactive              = errors.New("user status is inactive")
	ErrUserNotFound              = errors.New("user not found")
	ErrOtpRequired               = errors.New("otp required")
	ErrOtpInvalid                = errors.New("invalid otp")
	ErrOtpExpired                = errors.New("expired otp")
	ErrBadCredentials            = errors.New("invalid credentials")
	ErrInvalidCurrentPassword    = errors.New("invalid current password")
	ErrNewPasswordMatchesCurrent = errors.New("new password must not match current password")
	ErrTooFrequentPasswordChange = errors.New("password changed too frequently")
	ErrObjectIdCastFailed        = errors.New("failed to convert hex to objectId")
	ErrVerificationRequired      = errors.New("verification required")
)

// UserRepository defines the persistence methods required by UserService.
type UserRepository interface {
	Create(ctx context.Context, user entities.User) error
	ActiveAndRevokeToken(ctx context.Context, token string) error
	UpdateVerificationToken(ctx context.Context, userID primitive.ObjectID, token string) error
	SetHashPassowrd(ctx context.Context, userId primitive.ObjectID, passwordHash string, newTime, expiresAt time.Time) error
	SetLoginOtp(ctx context.Context, userId primitive.ObjectID, hash string, expiry time.Time) error
	ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FindUsersForExpiryNotification(ctx context.Context, daysUntilExpiry int, batchSize int, lastID string) ([]*entities.User, string, error)
	UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error
	FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Delete(ctx context.Context, userID primitive.ObjectID) error
}

// UserService contains business logic for user onboarding, login and account maintenance.
type UserService struct {
	r   UserRepository
	ms  MailSender
	jsc events.Client
	c   clock.Clock

	tr trace.Tracer
}

// NewUserService builds a UserService with repository and mail dependencies.
func NewUserService(r UserRepository, ms MailSender, jsc events.Client) *UserService {
	tr := otel.Tracer("user-service/user-service")
	s := UserService{r: r, ms: ms, c: clock.RealClock{}, jsc: jsc, tr: tr}

	return &s
}

// Register func, handles registration business logic such as username, email existence validation,
// sending verification mails.
func (s *UserService) Register(ctx context.Context, reqDto *dtos.UserRegistrationDto) error {
	registerCtx, registerSpan := s.tr.Start(ctx, "user.register")
	defer registerSpan.End()

	lookupCtx, lookupSpan := s.tr.Start(registerCtx, "user.register.lookup_unique")
	defer lookupSpan.End()
	// checks if the username is already taken,
	exists, err := s.r.ExistsByUsername(lookupCtx, reqDto.Username)
	if err != nil {
		lookupSpan.RecordError(err)
		return err
	}

	if exists {
		return ErrUsernameTaken
	}

	// checks if the email is already taken,
	exists, err = s.r.ExistsByEmail(lookupCtx, reqDto.Email)
	if err != nil {
		lookupSpan.RecordError(err)
		return err
	}

	if exists {
		return ErrEmailTaken
	}

	createCtx, createSpan := s.tr.Start(registerCtx, "user.register.create_user")
	defer createSpan.End()
	// if both the username and email are unique we convert the dto into the user entity,
	// the mapper method does all the heavy lifting and sets the default field values and
	// hashes the password,
	userEntity, err := mappers.ToUserEntity(reqDto)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(ctx, "error converting to user entity: %v", err)
		return err
	}

	// insert the user in the database,
	err = s.r.Create(createCtx, *userEntity)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrUsernameAlreadyTaken):
			err = ErrUsernameTaken
		case errors.Is(err, repositories.ErrEmailAlreadyTaken):
			err = ErrEmailTaken
		}
		createSpan.RecordError(err)
		logging.Errorf(ctx, "error creating user in database: %v", err)
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(registerCtx, 4*time.Second)
	defer cancel()

	eventCtx, eventSpan := s.tr.Start(timeoutCtx, "user.register.event")
	defer eventSpan.End()

	urp := toUserRegisteredPayload(userEntity.Username, userEntity.ID.Hex())

	err = retry.Do(
		func() error {
			return s.jsc.Publish(eventCtx, events.SUBJECT_USER_CREATED, urp)
		},
		retry.Attempts(3),
		retry.Delay(time.Second*1),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(eventCtx),
	)
	if err != nil {
		logging.Errorf(eventCtx, "failed to publish user created event: %v", err)
		eventSpan.RecordError(err)

		var errs []error

		errs = append(errs, err)

		rbCtx, rbSpan := s.tr.Start(registerCtx, "user.register.rollback")
		defer rbSpan.End()

		err := s.r.Delete(rbCtx, userEntity.ID)
		if err != nil {
			logging.Errorf(rbCtx, "rollback failed:  %v", err)
			rbSpan.RecordError(err)

			errs = append(errs, err)
		}

		return errors.Join(errs...)
	}

	_, mailSpan := s.tr.Start(registerCtx, "user.register.send_verification_email")
	defer mailSpan.End()

	if err := s.ms.SendAccountVerificationEmail(reqDto.Email, userEntity.EmailVerification.Token); err != nil {
		mailSpan.RecordError(err)
		logging.Errorf(registerCtx, "failed to send verification email: %v", err)
		return err
	}

	return nil
}

// VerifyAccount func, handles account verification business logic.
func (s *UserService) VerifyAccount(ctx context.Context, token string) error {
	ctx, span := s.tr.Start(ctx, "user.verify_account")
	defer span.End()

	if err := s.r.ActiveAndRevokeToken(ctx, token); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *UserService) resendVerification(ctx context.Context, user *entities.User) error {
	_, span := s.tr.Start(ctx, "user.resend_verification")
	defer span.End()

	token := uuid.NewString()
	if err := s.r.UpdateVerificationToken(ctx, user.ID, token); err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.ms.SendAccountVerificationEmail(user.Email, token); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// FindUsersForExpiryNotification func, finds users which password expires soon.
func (s *UserService) FindUsersForExpiryNotification(
	ctx context.Context,
	daysUntilExpiry int,
	batchSize int,
	lastID string,
) ([]*entities.User, string, error) {
	ctx, span := s.tr.Start(ctx, "user.find_users_for_expiry_notification")
	defer span.End()

	users, nextID, err := s.r.FindUsersForExpiryNotification(ctx, daysUntilExpiry, batchSize, lastID)
	if err != nil {
		span.RecordError(err)
		return nil, "", err
	}

	return users, nextID, nil
}

func (s *UserService) MarkExpiryNotificationSent(ctx context.Context, userID primitive.ObjectID) error {
	ctx, span := s.tr.Start(ctx, "user.mark_expiry_notification_sent")
	defer span.End()

	err := s.r.UpdateExpiryNotificationSentDate(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// Login func, validates credentials, checks account status, and issues a login OTP.
func (s *UserService) Login(ctx context.Context, loginDto *dtos.UserLoginDto) error {
	ctx, span := s.tr.Start(ctx, "user.login")
	defer span.End()

	lookupCtx, lookupSpan := s.tr.Start(ctx, "user.login.lookup_user")
	user, err := s.r.FindUserByEmail(lookupCtx, loginDto.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			lookupSpan.End()
			return ErrUserNotFound
		}

		lookupSpan.RecordError(err)
		lookupSpan.End()
		return err
	}
	lookupSpan.End()

	_, verifySpan := s.tr.Start(ctx, "user.login.verify_credentials")
	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		verifySpan.RecordError(err)
		verifySpan.End()
		return ErrBadCredentials
	}

	if user.AccountStatus == account.StatusInactive {
		verifySpan.End()
		if err := s.resendVerification(ctx, user); err != nil {
			return err
		}
		return ErrVerificationRequired
	}

	if s.c.Now().After(user.PasswordExpiresAt) {
		verifySpan.End()
		return ErrExpiredPassword
	}
	verifySpan.End()

	return s.sendLoginOtpToUser(ctx, user)
}

// CreateNewToken func, issues a signed JWT for the authenticated user.
func (s *UserService) CreateNewToken(ctx context.Context, user *entities.User) (string, error) {
	_, span := s.tr.Start(ctx, "user.create_token")
	defer span.End()

	now := s.c.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":      user.ID,
		"name":     strings.Join([]string{user.FirstName, user.LastName}, " "),
		"username": user.Username,
		"role":     user.Role,
		"status":   user.AccountStatus,
		"iat":      now.Unix(),
		"exp":      now.Add(15 * time.Minute).Unix(),
	})

	pk, err := utils.GetPrivateKey()
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	tokenString, err := token.SignedString(pk)

	return tokenString, err
}

// VerifyLoginOtp func, compares the provided OTP with the stored hash and clears it on success.
func (s *UserService) VerifyLoginOtp(ctx context.Context, dto *dtos.VerifyLoginOtpDto) (*entities.User, error) {
	ctx, span := s.tr.Start(ctx, "user.verify_login_otp")
	defer span.End()

	lookupCtx, lookupSpan := s.tr.Start(ctx, "user.verify_login_otp.lookup_user")
	user, err := s.r.FindUserByEmail(lookupCtx, dto.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			lookupSpan.End()
			return nil, ErrUserNotFound
		}

		lookupSpan.RecordError(err)
		lookupSpan.End()
		return nil, err
	}
	lookupSpan.End()

	validateCtx, validateSpan := s.tr.Start(ctx, "user.verify_login_otp.validate_code")
	if user.AccountStatus == account.StatusInactive {
		validateSpan.End()
		return nil, ErrUserInactive
	}

	if user.OTPCode.Hash == "" {
		validateSpan.End()
		return nil, ErrOtpInvalid
	}

	if s.c.Now().After(user.OTPCode.Expiry) {
		_ = s.r.ClearLoginOtp(validateCtx, user.ID)
		validateSpan.End()
		return nil, ErrOtpExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.OTPCode.Hash), []byte(dto.Code)); err != nil {
		validateSpan.RecordError(err)
		validateSpan.End()
		return nil, ErrOtpInvalid
	}
	validateSpan.End()

	clearCtx, clearSpan := s.tr.Start(ctx, "user.verify_login_otp.clear_otp")
	_ = s.r.ClearLoginOtp(clearCtx, user.ID)
	clearSpan.End()
	return user, nil
}

// ResendLoginOtp resends the OTP code to the user's email for login verification.
func (s *UserService) ResendLoginOtp(ctx context.Context, email string) error {
	ctx, span := s.tr.Start(ctx, "user.resend_login_otp")
	defer span.End()

	user, err := s.r.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return ErrUserNotFound
		}

		span.RecordError(err)
		return err
	}

	if user.AccountStatus == account.StatusInactive {
		return ErrUserInactive
	}

	return s.sendLoginOtpToUser(ctx, user)
}

// sendLoginOtpToUser is a private helper that generates and sends OTP to a user.
func (s *UserService) sendLoginOtpToUser(ctx context.Context, user *entities.User) error {
	ctx, span := s.tr.Start(ctx, "user.login.issue_otp")
	defer span.End()

	otp, err := auth.GenerateOTP()
	if err != nil {
		span.RecordError(err)
		return err
	}

	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)

	// TODO: make time NOT be hardcoded
	if err := s.r.SetLoginOtp(ctx, user.ID, string(otpHash), s.c.Now().Add(5*time.Minute)); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.ms.SendLoginOtp(user.Email, otp); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *UserService) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	ctx, span := s.tr.Start(ctx, "user.find_by_id")
	defer span.End()

	user, err := s.r.FindUserByID(ctx, id)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		span.RecordError(err)
	}

	return user, err
}

// EmailExists checks if an email is already registered.
func (s *UserService) EmailExists(ctx context.Context, email string) (bool, error) {
	ctx, span := s.tr.Start(ctx, "user.email_exists")
	defer span.End()

	exists, err := s.r.ExistsByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
	}
	return exists, err
}

func (s *UserService) ChangePassword(ctx context.Context, dto *dtos.ChangePasswordDto) error {
	// Extract user ID before creating spans
	userIdHexString := middlewares.GetUserIdFromContext(ctx)
	if userIdHexString == "" {
		return ErrObjectIdCastFailed
	}

	ctx, span := s.tr.Start(ctx, "user.change_password")
	defer span.End()

	userObjectId, err := primitive.ObjectIDFromHex(userIdHexString)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	_, lookupSpan := s.tr.Start(ctx, "user.change_password.lookup_user")
	user, err := s.r.FindUserByID(ctx, userObjectId)
	if err != nil {
		// User not found error should not appear here as the user is authenticated
		// and user id is extracted from the token. But just in case, we log and return.
		lookupSpan.RecordError(err)
		lookupSpan.End()
		return err
	}
	lookupSpan.End()

	_, passwordSpan := s.tr.Start(ctx, "user.change_password.validate_and_set")
	err = auth.CompareHashAndPassword(user.Password, dto.CurrentPassword)
	if err != nil {
		passwordSpan.RecordError(err)
		passwordSpan.End()
		return ErrInvalidCurrentPassword
	}

	if dto.CurrentPassword == dto.NewPassword {
		passwordSpan.End()
		return ErrNewPasswordMatchesCurrent
	}

	if user.PasswordLastChanged.Compare(s.c.Now().Add(-24*time.Hour)) >= 0 {
		passwordSpan.End()
		return ErrTooFrequentPasswordChange
	}

	hashedPassword, err := auth.HashPassword(dto.NewPassword)
	if err != nil {
		passwordSpan.RecordError(err)
		passwordSpan.End()
		return err
	}

	newTime := s.c.Now()
	expiresAt := newTime.Add(60 * 24 * time.Hour)

	if err := s.r.SetHashPassowrd(ctx, user.ID, hashedPassword, newTime, expiresAt); err != nil {
		passwordSpan.RecordError(err)
		passwordSpan.End()
		return err
	}

	passwordSpan.End()
	return nil
}

func toUserRegisteredPayload(username string, userID string) events.UserRegistrationPayload {
	return events.UserRegistrationPayload{
		UserID:   userID,
		Username: username,
	}
}
