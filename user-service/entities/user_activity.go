package entities

import "time"

type ActivityType string

const (
	ActivityListen             ActivityType = "LISTEN"
	ActivitySubscriptionCreate ActivityType = "SUBSCRIPTION_CREATED"
	ActivitySubscriptionDelete ActivityType = "SUBSCRIPTION_DELETED"
	ActivityRatingCreate       ActivityType = "RATING_CREATED"
	ActivityRatingUpdate       ActivityType = "RATING_UPDATED"
)

type UserActivity struct {
	EventID     string       `bson:"event_id" json:"event_id"`
	UserID      string       `bson:"user_id" json:"user_id"`
	Type        ActivityType `bson:"type" json:"type"`
	EntityID    string       `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	EntityName  string       `bson:"entity_name,omitempty" json:"entity_name,omitempty"`
	EntityType  string       `bson:"entity_type,omitempty" json:"entity_type,omitempty"`
	RatingValue *int         `bson:"rating_value,omitempty" json:"rating_value,omitempty"`
	OccurredAt  time.Time    `bson:"occurred_at" json:"occurred_at"`
	CreatedAt   time.Time    `bson:"created_at" json:"created_at"`
}
