package mappers

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrPasswordHashing is returned when bcrypt hashing fails.
var ErrPasswordHashing = errors.New("an error has occurred while hashing password")

// ToUserEntity maps a registration DTO into a fully initialized User entity.
func ToUserEntity(u *dtos.UserRegistrationDto) (*entities.User, error) {
	now := time.Now()
	expiryDate := time.Now().Add(60 * 24 * time.Hour)

	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		return nil, ErrPasswordHashing
	}

	return &entities.User{
		ID:                  primitive.NewObjectID(),
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Email:               u.Email,
		Username:            u.Username,
		Password:            hashedPassword,
		AccountStatus:       account.StatusInactive,
		Role:                account.RoleMember,
		CreatedAt:           now,
		UpdatedAt:           now,
		PasswordLastChanged: now,
		PasswordExpiresAt:   expiryDate,
		EmailVerification: entities.EmailVerification{
			Token: uuid.New().String(),
			Type:  entities.AccountVerification,
		},
	}, nil
}
