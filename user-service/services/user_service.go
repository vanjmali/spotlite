package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/clock"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
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
	ErrTooFrequentPasswordChange = errors.New("password changed too frequently")
	ErrObjectIdCastFailed        = errors.New("failed to convert hex to objectId")
)

// UserRepository defines the persistence methods required by UserService.
type UserRepository interface {
	Create(ctx context.Context, user entities.User) error
	ActiveAndRevokeToken(ctx context.Context, token string) error
	SetHashPassowrd(ctx context.Context, userId primitive.ObjectID, passwordHash string, newTime, expiresAt time.Time) error
	SetLoginOtp(ctx context.Context, userId primitive.ObjectID, hash string, expiry time.Time) error
	ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FindUsersForExpiryNotification(ctx context.Context, daysUntilExpiry int, batchSize int, lastID string) ([]*entities.User, string, error)
	UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error
	FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// UserService contains business logic for user onboarding, login and account maintenance.
type UserService struct {
	r  UserRepository
	ms MailSender
	c  clock.Clock

	tr trace.Tracer
}

// NewUserService builds a UserService with repository and mail dependencies.
func NewUserService(r UserRepository, ms MailSender) *UserService {
	tr := otel.Tracer("user-service/user-service")
	s := UserService{r: r, ms: ms, c: clock.RealClock{}, tr: tr}

	return &s
}

// Register func, handles registration business logic such as username, email existence validation,
// sending verification mails.
func (s *UserService) Register(ctx context.Context, reqDto *dtos.UserRegistrationDto) error {
	ctx, span := s.tr.Start(ctx, "user.register")
	defer span.End()

	lookupCtx, lookupSpan := s.tr.Start(ctx, "user.register.lookup_unique")
	// checks if the username is already taken,
	exists, err := s.r.ExistsByUsername(lookupCtx, reqDto.Username)
	if err != nil {
		lookupSpan.RecordError(err)
		lookupSpan.End()
		log.Printf("Error checking if username exists: %v", err)
		return err
	}
	if exists {
		lookupSpan.End()
		log.Printf("Username already taken: %s", reqDto.Username)
		return ErrUsernameTaken
	}

	// checks if the email is already taken,
	exists, err = s.r.ExistsByEmail(lookupCtx, reqDto.Email)
	if err != nil {
		lookupSpan.RecordError(err)
		lookupSpan.End()
		log.Printf("Error checking if email exists: %v", err)
		return err
	}
	if exists {
		lookupSpan.End()
		log.Printf("Email already taken: %s", reqDto.Email)
		return ErrEmailTaken
	}
	lookupSpan.End()

	createCtx, createSpan := s.tr.Start(ctx, "user.register.create_user")
	// if both the username and email are unique we convert the dto into the user entity,
	// the mapper method does all the heavy lifting and sets the default field values and
	// hashes the password,
	userEntity, err := mappers.ToUserEntity(reqDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to user entity: %v", err)
		return err
	}

	_, mailSpan := s.tr.Start(ctx, "user.register.send_verification_email")
	// sends account verification email BEFORE saving to database
	// if email fails, we don't save the user
	if err := s.ms.SendAccountVerificationEmail(reqDto.Email, userEntity.EmailVerification.Token); err != nil {
		mailSpan.RecordError(err)
		mailSpan.End()
		log.Printf("Failed to send verification email: %v", err)
		return err
	}
	mailSpan.End()

	// insert the user in the database,
	err = s.r.Create(createCtx, *userEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating user in database: %v", err)
		return err
	}
	createSpan.End()

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
		lookupSpan.RecordError(err)
		lookupSpan.End()
		return ErrUserNotFound
	}
	lookupSpan.End()

	_, verifySpan := s.tr.Start(ctx, "user.login.verify_credentials")
	if user.AccountStatus == account.StatusInactive {
		verifySpan.End()
		return ErrUserInactive
	}

	if s.c.Now().After(user.PasswordExpiresAt) {
		verifySpan.End()
		return ErrExpiredPassword
	}

	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		verifySpan.RecordError(err)
		verifySpan.End()
		return ErrBadCredentials
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
		span.RecordError(err)
		return ErrUserNotFound
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
	if err != nil {
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
	ctx, span := s.tr.Start(ctx, "user.change_password")
	defer span.End()

	lookupCtx, lookupSpan := s.tr.Start(ctx, "user.change_password.lookup_user")
	userIdHexString := middlewares.GetUserIdFromContext(lookupCtx)

	userObjectId, err := primitive.ObjectIDFromHex(userIdHexString)
	if err != nil {
		lookupSpan.RecordError(err)
		lookupSpan.End()
		return ErrObjectIdCastFailed
	}

	user, err := s.r.FindUserByID(ctx, userObjectId)
	if err != nil {
		lookupSpan.RecordError(err)
		lookupSpan.End()
		return err
	}
	lookupSpan.End()

	_, passwordSpan := s.tr.Start(ctx, "user.change_password.validate_and_set")
	if user.PasswordLastChanged.Compare(s.c.Now().Add(-24*time.Hour)) >= 0 {
		passwordSpan.End()
		return ErrTooFrequentPasswordChange
	}

	err = auth.CompareHashAndPassword(user.Password, dto.CurrentPassword)
	if err != nil {
		passwordSpan.RecordError(err)
		passwordSpan.End()
		return ErrInvalidCurrentPassword
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
