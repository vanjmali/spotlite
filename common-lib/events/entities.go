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
	ANALYTICS_STREAM     = "ANALYTICS"

	SUBJECT_ENTITY_CREATED = "content.created"
	ENTITY_CREATE_DURABLE  = "ENTITY_CREATOR"

	SUBJECT_ENTITY_UPDATED = "content.updated"
	ENTITY_UPDATE_DURABLE  = "ENTITY_UPDATER"

	SUBJECT_SUBSCRIBER_BATCH = "subscribers.batch.process"
	SUB_DURABLE              = "SUBSCRIBER_PROCESSOR"

	// Analytics event subjects and durables
	SUBJECT_SONG_PLAYED          = "analytics.song_played"
	SONG_PLAYED_DURABLE          = "SONG_PLAYED_PROCESSOR"
	SUBJECT_RATING_CREATED       = "analytics.rating_created"
	RATING_CREATED_DURABLE       = "RATING_CREATED_PROCESSOR"
	SUBJECT_RATING_UPDATED       = "analytics.rating_updated"
	RATING_UPDATED_DURABLE       = "RATING_UPDATED_PROCESSOR"
	SUBJECT_RATING_DELETED       = "analytics.rating_deleted"
	RATING_DELETED_DURABLE       = "RATING_DELETED_PROCESSOR"
	SUBJECT_SUBSCRIPTION_CREATED = "analytics.subscription_created"
	SUBSCRIPTION_CREATED_DURABLE = "SUBSCRIPTION_CREATED_PROCESSOR"
	SUBJECT_SUBSCRIPTION_DELETED = "analytics.subscription_deleted"
	SUBSCRIPTION_DELETED_DURABLE = "SUBSCRIPTION_DELETED_PROCESSOR"
	SUBJECT_SONG_DELETED         = "analytics.song_deleted"
	SONG_DELETED_DURABLE         = "SONG_DELETED_PROCESSOR"
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

// Analytics event payloads
type SongPlayedEventPayload struct {
	UserID     string    `json:"user_id"`
	SongID     string    `json:"song_id"`
	ArtistID   string    `json:"artist_id"`
	AlbumID    string    `json:"album_id"`
	GenreID    string    `json:"genre_id"`
	DurationMS int       `json:"duration_ms"`
	PlayedAt   time.Time `json:"played_at"`
	EventID    string    `json:"event_id"`
}

type RatingCreatedEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	EventID   string    `json:"event_id"`
}

type RatingUpdatedEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	OldRating int       `json:"old_rating"`
	NewRating int       `json:"new_rating"`
	UpdatedAt time.Time `json:"updated_at"`
	EventID   string    `json:"event_id"`
}

type RatingDeletedEventPayload struct {
	UserID        string    `json:"user_id"`
	SongID        string    `json:"song_id"`
	DeletedRating int       `json:"deleted_rating"`
	DeletedAt     time.Time `json:"deleted_at"`
	EventID       string    `json:"event_id"`
}

type SubscriptionCreatedEventPayload struct {
	UserID           string    `json:"user_id"`
	SubscriptionType string    `json:"subscription_type"`
	TargetID         string    `json:"target_id"`
	CreatedAt        time.Time `json:"created_at"`
	EventID          string    `json:"event_id"`
}

type SubscriptionDeletedEventPayload struct {
	UserID           string    `json:"user_id"`
	SubscriptionType string    `json:"subscription_type"`
	TargetID         string    `json:"target_id"`
	DeletedAt        time.Time `json:"deleted_at"`
	EventID          string    `json:"event_id"`
}

type SongDeletedEventPayload struct {
	SongID    string    `json:"song_id"`
	DeletedAt time.Time `json:"deleted_at"`
	EventID   string    `json:"event_id"`
}
