package mappers

import (
	"time"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToUserEntity(u *dtos.UserRegistrationDto) *entities.User {
	now := time.Now()
	expiryDate := time.Now().Add(60 * 24 * time.Hour)

	hashedPassword, _ := auth.HashPassword(u.Password)

	return &entities.User{
		ID:                  primitive.NewObjectID(),
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Email:               u.Email,
		Username:            u.Username,
		Password:            hashedPassword,
		AccountStatus:       entities.StatusInactive,
		Role:                entities.RoleMember,
		CreatedAt:           now,
		UpdatedAt:           now,
		PasswordLastChanged: now,
		PasswordExpiresAt:   expiryDate,
	}
}
