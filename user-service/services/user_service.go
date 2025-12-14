package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/mappers"
	"github.com/vanjmali/spotlite/user-service/repositories"
)

var (
	ErrUsernameTaken = errors.New("username is already taken")
	ErrEmailTaken    = errors.New("email is already taken")
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
	return s.r.ActiveAndRevokeToken(ctx, token)
}
