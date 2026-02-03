package entities

import (
	"time"

	"github.com/vanjmali/spotlite/common-lib/subscription"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID           primitive.ObjectID            `bson:"_id,omitempty" json:"id"`
	SubscriberID primitive.ObjectID            `bson:"subscriber_id,omitempty" json:"subscriber_id"`
	EntityID     primitive.ObjectID            `bson:"entity_id,omitempty" json:"entity_id"`
	Type         subscription.SubscriptionType `bson:"sub_type,omitempty" json:"sub_type"`
	SubscribedAt time.Time                     `bson:"subscribed_at,omitempty" json:"subscribed_at"`
	EntityName   string                        `bson:"entity_name,omitempty" json:"entity_name"`
}
