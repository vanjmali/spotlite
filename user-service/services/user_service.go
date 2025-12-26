package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUsernameTaken indicates the supplied username already exists.
	ErrUsernameTaken = errors.New("username is already taken")
	// ErrEmailTaken indicates the supplied email already exists.
	ErrEmailTaken = errors.New("email is already taken")
	// ErrExpiredPassword signals that the user's password has expired.
	ErrExpiredPassword = errors.New("your password is expired")
	// ErrUserInactive marks an inactive account status.
	ErrUserInactive = errors.New("user status is inactive")
	// ErrUserNotFound indicates the user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrOtpRequired indicates login requires an OTP code.
	ErrOtpRequired = errors.New("otp required")
	// ErrOtpInvalid indicates a provided OTP is wrong.
	ErrOtpInvalid = errors.New("invalid otp")
	// ErrOtpExpired indicates the OTP is no longer valid.
	ErrOtpExpired = errors.New("expired otp")
	// ErrBadCredentials indicates the credentials are invalid.
	ErrBadCredentials = errors.New("invalid credentials")
)

// UserService contains business logic for user onboarding, login and account maintenance.
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
		log.Printf("Error checking if username exists: %v", err)
		return err
	}
	if exists {
		log.Printf("Username already taken: %s", reqDto.Username)
		return ErrUsernameTaken
	}

	// checks if the email is already taken,
	exists, err = s.r.ExistsByEmail(ctx, reqDto.Email)
	if err != nil {
		log.Printf("Error checking if email exists: %v", err)
		return err
	}
	if exists {
		log.Printf("Email already taken: %s", reqDto.Email)
		return ErrEmailTaken
	}

	// if both the username and email are unique we convert the dto into the user entity,
	// the mapper method does all the heavy lifting and sets the default field values and
	// hashes the password,
	userEntity, err := mappers.ToUserEntity(reqDto)
	if err != nil {
		log.Printf("Error converting to user entity: %v", err)
		return err
	}

	// sends account verification email BEFORE saving to database
	// if email fails, we don't save the user
	if err := s.ms.sendAccountVerificationEmail(reqDto.Email, userEntity.EmailVerification.Token); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		return err
	}

	// insert the user in the database,
	err = s.r.Create(ctx, *userEntity)
	if err != nil {
		log.Printf("Error creating user in database: %v", err)
		return err
	}

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

	if user.AccountStatus == account.StatusInactive {
		return ErrUserInactive
	}

	if time.Now().After(user.PasswordExpiresAt) {
		return ErrExpiredPassword
	}

	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		return ErrBadCredentials
	}

	return s.sendLoginOtpToUser(ctx, user)
}

// CreateNewToken issues a signed JWT for the authenticated user.
func (s *UserService) CreateNewToken(ctx context.Context, user *entities.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":      user.ID,
		"name":     strings.Join([]string{user.FirstName, user.LastName}, " "),
		"username": user.Username,
		"role":     user.Role,
		"status":   user.AccountStatus,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	})

	pk, err := utils.GetPrivateKey()
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(pk)

	return tokenString, err
}

// VerifyLoginOtp compares the provided OTP with the stored hash and clears it on success.
func (s *UserService) VerifyLoginOtp(ctx context.Context, dto *dtos.VerifyLoginOtpDto) (*entities.User, error) {
	user, err := s.r.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}

	if user.AccountStatus == account.StatusInactive {
		return nil, ErrUserInactive
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

// ResendLoginOtp resends the OTP code to the user's email for login verification.
func (s *UserService) ResendLoginOtp(ctx context.Context, email string) error {
	user, err := s.r.FindUserByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	if user.AccountStatus == account.StatusInactive {
		return ErrUserInactive
	}

	return s.sendLoginOtpToUser(ctx, user)
}

// sendLoginOtpToUser is a private helper that generates and sends OTP to a user.
func (s *UserService) sendLoginOtpToUser(ctx context.Context, user *entities.User) error {
	otp, err := auth.GenerateOTP()
	if err != nil {
		return err
	}

	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)

	// TODO: make time NOT be hardcoded
	if err := s.r.SetLoginOtp(ctx, user.ID, string(otpHash), time.Now().Add(5*time.Minute)); err != nil {
		return err
	}
	if err := s.ms.SendLoginOtp(user.Email, otp); err != nil {
		return err
	}

	return nil
}

func (s *UserService) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	return s.r.FindUserByID(ctx, id)
}

// EmailExists checks if an email is already registered
func (s *UserService) EmailExists(ctx context.Context, email string) (bool, error) {
	return s.r.ExistsByEmail(ctx, email)
}
