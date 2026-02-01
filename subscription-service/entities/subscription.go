package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionType string

const (
	GenreSubscription  SubscriptionType = "genre"
	ArtistSubscription SubscriptionType = "artist"
)

type Subscription struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SubscriberID primitive.ObjectID `bson:"subscriber_id,omitempty" json:"subscriber_id"`
	EntityID     primitive.ObjectID `bson:"entity_id,omitempty" json:"entity_id"`
	Type         SubscriptionType   `bson:"sub_type,omitempty" json:"sub_type"`
	SubscribedAt time.Time          `bson:"subscribed_at,omitempty" json:"subscribed_at"`
	EntityName   string             `bson:"entity_name,omitempty" json:"entity_name"`
}
