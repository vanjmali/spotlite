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
	USERS_STREAM         = "USERS"
	SONGS_STREAM         = "SONGS"
	GENRES_STREAM        = "GENRES"
	ARTISTS_STREAM       = "ARTISTS"

	SUBJECT_ENTITY_CREATED = "content.created"
	ENTITY_CREATE_DURABLE  = "ENTITY_CREATOR"

	SUBJECT_ENTITY_UPDATED = "content.updated"
	ENTITY_UPDATE_DURABLE  = "ENTITY_UPDATER"

	SUBJECT_SUBSCRIBER_BATCH = "subscribers.batch.process"
	SUB_DURABLE              = "SUBSCRIBER_PROCESSOR"

	SUBJECT_USER_CREATED = "user.created"
	USER_DURABLE         = "USER_PROCESSOR"

	SUBJECT_GENRE_CREATED    = "genre.created"
	SUBJECT_GENRE_SUBSCRIBED = "genre.subscription.created"
	GENRE_DURABLE            = "GENRE_PROCESSOR"

	SUBJECT_SONG_CREATED = "songs.created"
	SUBJECT_SONG_RATED   = "songs.rating.created"
	SONG_DURABLE         = "SONG_PROCESSOR"

	SUBJECT_ARTIST_CREATED = "artist.created"
	ARTIST_DURABLE         = "ARTIST_PROCESSOR"
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
