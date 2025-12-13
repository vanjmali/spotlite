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
)

type UserService struct {
	r *repositories.UserRepository
}

func NewUserService(r repositories.UserRepository) *UserService {
	s := UserService{r: &r}

	return &s
}

func (s *UserService) Register(ctx context.Context, reqDto *dtos.UserRegistrationDto) error {
	exists, err := s.r.ExistsByUsername(ctx, reqDto.Username)

	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameTaken
	}

	exists, err = s.r.ExistsByEmail(ctx, reqDto.Email)

	if err != nil {
		return err
	}
	if exists {
		return ErrEmailTaken
	}

	userEntity := mappers.ToUserEntity(reqDto)
	err = s.r.Create(ctx, *userEntity)
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
