package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/models"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
)

var (
	ErrUsernameTaken = errors.New("username is already taken")
	ErrEmailTaken    = errors.New("email is already taken")
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

func (s *UserService) Login(ctx context.Context, loginDto *dtos.UserLoginDto) (*models.User, error) {
	hashedPassword, err := auth.HashPassword(loginDto.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.r.FindUserByEmailAndPassword(ctx, loginDto.Email, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}
