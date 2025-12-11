package mappers

import (
	"time"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToUserEntity(u *dtos.UserRegistrationDto) *models.User {
	now := time.Now()
	expiryDate := time.Now().Add(60 * 24 * time.Hour)

	return &models.User{
		ID:                  primitive.NewObjectID(),
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Email:               u.Email,
		Username:            u.Username,
		Password:            u.Password,
		AccountStatus:       models.StatusInactive,
		Role:                models.RoleMember,
		CreatedAt:           now,
		UpdatedAt:           now,
		PasswordLastChanged: now,
		PasswordExpiresAt:   expiryDate,
	}
}
