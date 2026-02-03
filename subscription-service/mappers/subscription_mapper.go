package mappers

import (
	"errors"
	"time"

	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrSubscriptionMapping = errors.New("an error has occurred while processing subscription request")

func ToSubscriptionEntity(req *dtos.CreateSubscriptionDto, subscriberIDstr string, entityName string) (*entities.Subscription, error) {
	now := time.Now()

	subscriberID, err := primitive.ObjectIDFromHex(subscriberIDstr)
	if err != nil {
		return nil, ErrSubscriptionMapping
	}

	entityId, err := primitive.ObjectIDFromHex(req.EntityID)
	if err != nil {
		return nil, ErrSubscriptionMapping
	}

	return &entities.Subscription{
		ID:           primitive.NewObjectID(),
		SubscriberID: subscriberID,
		EntityID:     entityId,
		Type:         req.Type,
		SubscribedAt: now,
		EntityName:   entityName,
	}, nil
}
