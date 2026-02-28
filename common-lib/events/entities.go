package events

import (
	"time"
)

type EntityType string

const (
	ArtistType EntityType = "ARTIST"
	AlbumType  EntityType = "ALBUM"

	CONTENT_STREAM       = "CONTENT"
	SUBSCRIPTIONS_STREAM = "SUBSCRIPTIONS"
	RATINGS_STREAM       = "RATINGS"
	LISTENS_STREAM       = "LISTENS"

	SUBJECT_ENTITY_CREATED = "content.created"
	ENTITY_CREATE_DURABLE  = "ENTITY_CREATOR"

	SUBJECT_ENTITY_UPDATED = "content.updated"
	ENTITY_UPDATE_DURABLE  = "ENTITY_UPDATER"

	SUBJECT_SUBSCRIBER_BATCH = "subscribers.batch.process"
	SUB_DURABLE              = "SUBSCRIBER_PROCESSOR"

	SUBJECT_RATING_CREATED = "rating.created"
	SUBJECT_RATING_UPDATED = "rating.updated"
	SUBJECT_RATING_DELETED = "rating.deleted"
	RATING_DURABLE         = "RATING_PROCESSOR"

	SUBJECT_LISTEN_CREATED = "listen.created"
	LISTEN_DURABLE         = "LISTEN_PROCESSOR"

	SUBJECT_SUBSCRIPTION_CREATED = "subscription.created"
	SUBJECT_SUBSCRIPTION_DELETED = "subscription.deleted"
	SUBSCRIPTION_DURABLE         = "SUBSCRIPTION_PROCESSOR"
)

type SubscriptionEntityType string

const (
	SubscriptionEntityArtist SubscriptionEntityType = "ARTIST"
	SubscriptionEntityGenre  SubscriptionEntityType = "GENRE"
)

type EntityCreatedEventPayload struct {
	TargetIDs  []string   `json:"target_ids"`
	CreatedAt  time.Time  `json:"created_at"`
	EntityID   string     `json:"entity_id"`
	EntityName string     `json:"entity_name"`
	EntityType EntityType `json:"entity_type"`
	EventID    string     `json:"event_id"`
}

type EntityUpdatedEventPayload struct {
	EntityID   string `json:"entity_id"`
	EntityName string `json:"entity_name"`
}

type SubscribersBatchEventPayload struct {
	EntityID      string     `json:"entity_id"`
	EntityName    string     `json:"entity_name"`
	EntityType    EntityType `json:"entity_type"`
	CreatedAt     time.Time  `json:"created_at"`
	SubscriberIDs []string   `json:"subscriber_ids"`
	EventID       string     `json:"event_id"`
}

type RatingEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	Rating    int       `json:"rating"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ListenEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type SubscriptionEventPayload struct {
	UserID     string                 `json:"user_id"`
	EntityID   string                 `json:"entity_id"`
	EntityType SubscriptionEntityType `json:"entity_type"`
	EventID    string                 `json:"event_id"`
	CreatedAt  time.Time              `json:"created_at"`
}
