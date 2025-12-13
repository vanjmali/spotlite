package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
)

// TODO: Convert to .env
const key string = "superSecretPassword123"

var hmacSampleSecret []byte = []byte(key)

var (
	ErrUsernameTaken   = errors.New("username is already taken")
	ErrEmailTaken      = errors.New("email is already taken")
	ErrExpiredPassword = errors.New("your password is expired")
	ErrUserInnactive   = errors.New("user status is innactive")
)

type UserService struct {
	r  *repositories.UserRepository
	ms *MailService
}

func NewUserService(r repositories.UserRepository, ms MailService) *UserService {
	s := UserService{r: &r, ms: &ms}

	return &s
}

/*
Register func, handles registration business logic such as username, email existence validation,
sending verification mails
*/
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

	/**
		if both the username and email are unique we convert the dto into the user entity,
		the mapper method does all the heavy lifting and sets the default field values and
		hashes the password,
	**/
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
	s.ms.sendAccountVerificationEmail(reqDto.Email, userEntity.Token.Content)
	return nil
}

// VerifyAccount func, handles account verification business logic
func (s *UserService) VerifyAccount(ctx context.Context, token string) error {
	err := s.r.ActiveAndRevokeToken(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, loginDto *dtos.UserLoginDto) (*entities.User, error) {

	user, err := s.r.FindUserByEmail(ctx, loginDto.Email)
	if err != nil {
		return nil, err
	}

	if user.AccountStatus == entities.StatusInactive {
		return nil, ErrUserInnactive
	}

	if time.Now().After(user.PasswordExpiresAt) {
		return nil, ErrExpiredPassword
	}

	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateNewToken(ctx context.Context, user *entities.User) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"name":     strings.Join([]string{user.FirstName, user.LastName}, " "),
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(hmacSampleSecret)
	return tokenString, err
}
