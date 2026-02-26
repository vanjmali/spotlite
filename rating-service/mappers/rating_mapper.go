package mappers

import (
	"errors"
	"time"

	"github.com/vanjmali/spotlite/common-lib/types"
	"github.com/vanjmali/spotlite/rating-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrRatingMapping = errors.New("an error has occurred while processing rating request")

// ToRatingEntity converts songID, userID, ratingValue, and username into a Rating entity.
func ToRatingEntity(songIDStr string, userIDStr string, ratingValue int, username string) (*entities.Rating, error) {
	now := time.Now()

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrRatingMapping
	}

	songID, err := primitive.ObjectIDFromHex(songIDStr)
	if err != nil {
		return nil, ErrRatingMapping
	}

	return &entities.Rating{
		ID:        primitive.NewObjectID(),
		SongID:    songID,
		UserID:    userID,
		Value:     ratingValue,
		Username:  username,
		CreatedAt: now,
		IsEdited:  false,
		Status:    types.StatusActive,
	}, nil
}
