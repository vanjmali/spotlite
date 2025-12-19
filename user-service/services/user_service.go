package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var hmacSampleSecret = []byte(utils.MustGetEnv("APP_JWT_SECRET"))
var loginOtpTTL = utils.MustGetDurationEnv("APP_LOGIN_OTP_TTL_MINUTES", time.Minute)

var (
	ErrUsernameTaken             = errors.New("username is already taken")
	ErrEmailTaken                = errors.New("email is already taken")
	ErrExpiredPassword           = errors.New("your password is expired")
	ErrUserInnactive             = errors.New("user status is innactive")
	ErrOtpRequired               = errors.New("otp required")
	ErrOtpInvalid                = errors.New("invalid otp")
	ErrOtpExpired                = errors.New("expired otp")
	ErrBadCredentials            = errors.New("invalid credentials")
	ErrInvalidCurrentPassword    = errors.New("current password is incorrect")
	ErrTooFrequentPasswordChange = errors.New("password changed too frequent")
)

type UserService struct {
	r  *repositories.UserRepository
	ms *MailService
}

// NewUserService builds a UserService with repository and mail dependencies.
func NewUserService(r repositories.UserRepository, ms MailService) *UserService {
	s := UserService{r: &r, ms: &ms}

	return &s
}

// Register func, handles registration business logic such as username, email existence validation,
// sending verification mails.
func (s *UserService) Register(ctx context.Context, reqDto *dtos.UserRegistrationDto) error {
	// checks if the username is already taken,
	exists, err := s.r.ExistsByUsername(ctx, reqDto.Username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameTaken
	}

	// checks if the email is already taken,
	exists, err = s.r.ExistsByEmail(ctx, reqDto.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailTaken
	}

	// if both the username and email are unique we convert the dto into the user entity,
	// the mapper method does all the heavy lifting and sets the default field values and
	// hashes the password,
	userEntity, err := mappers.ToUserEntity(reqDto)
	if err != nil {
		return err
	}

	// insert the user in the database,
	err = s.r.Create(ctx, *userEntity)
	if err != nil {
		return err
	}

	// sends account verification email,
	s.ms.sendAccountVerificationEmail(reqDto.Email, userEntity.EmailVerification.Token)
	return nil
}

// VerifyAccount func, handles account verification business logic.
func (s *UserService) VerifyAccount(ctx context.Context, token string) error {
	return s.r.ActiveAndRevokeToken(ctx, token)
}

// Login validates credentials, checks account status, and issues a login OTP.
func (s *UserService) Login(ctx context.Context, loginDto *dtos.UserLoginDto) error {
	user, err := s.r.FindUserByEmail(ctx, loginDto.Email)
	if err != nil {
		return err
	}

	if user.AccountStatus == entities.StatusInactive {
		return ErrUserInnactive
	}

	if time.Now().After(user.PasswordExpiresAt) {
		return ErrExpiredPassword
	}

	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		return ErrBadCredentials
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		return err
	}

	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)

	if err := s.r.SetLoginOtp(ctx, user.ID, string(otpHash), time.Now().Add(loginOtpTTL)); err != nil {
		return err
	}
	if err := s.ms.SendLoginOtp(user.Email, otp); err != nil {
		return err
	}

	return nil
}

// CreateNewToken issues a signed JWT for the authenticated user.
func (s *UserService) CreateNewToken(ctx context.Context, user *entities.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"name":     strings.Join([]string{user.FirstName, user.LastName}, " "),
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString(hmacSampleSecret)
	return tokenString, err
}

// VerifyLoginOtp compares the provided OTP with the stored hash and clears it on success.
func (s *UserService) VerifyLoginOtp(ctx context.Context, dto *dtos.VerifyLoginOtpDto) (*entities.User, error) {
	user, err := s.r.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}

	if user.AccountStatus == entities.StatusInactive {
		return nil, ErrUserInnactive
	}

	if user.OTPCode.Hash == "" {
		return nil, ErrOtpInvalid
	}

	if time.Now().After(user.OTPCode.Expiry) {
		_ = s.r.ClearLoginOtp(ctx, user.ID)
		return nil, ErrOtpExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.OTPCode.Hash), []byte(dto.Code)); err != nil {
		return nil, ErrOtpInvalid
	}

	_ = s.r.ClearLoginOtp(ctx, user.ID)
	return user, nil
}

func (s *UserService) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	return s.r.FindUserByID(ctx, id)
}

func (s *UserService) ChangePassword(ctx context.Context, dto *dtos.ChangePasswordDto) error {
	// TODO: remove email, use middleware when it's implemented
	user, err := s.r.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		return err
	}

	if user.PasswordLastChanged.Compare(time.Now().Add(-24*time.Hour)) >= 0 {
		return ErrTooFrequentPasswordChange
	}

	err = auth.CompareHashAndPassword(user.Password, dto.CurrentPassword)
	if err != nil {
		return ErrInvalidCurrentPassword
	}

	hashedPassword, err := auth.HashPassword(dto.NewPassword)
	if err != nil {
		return err
	}

	newTime := time.Now()

	expiresAt := newTime.Add(60 * 24 * time.Hour)

	if err := s.r.SetHashPassowrd(ctx, user.ID, hashedPassword, newTime, expiresAt); err != nil {
		return err
	}

	return err
}
