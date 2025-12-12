package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/mappers"
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

func (s *UserService) Login(ctx context.Context, loginDto *dtos.UserLoginDto) (*entities.User, error) {

	user, err := s.r.FindUserByEmail(ctx, loginDto.Email)

	if err != nil {
		return nil, err
	}
	err = auth.CompareHashAndPassword(user.Password, loginDto.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}
