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
	GENRE_CREATE_DURABLE     = "GENRE_CREATE_PROCESSOR"
	GENRE_SUB_DURABLE        = "GENRE_SUB_PROCESSOR"

	SUBJECT_SONG_CREATED = "songs.created"
	SUBJECT_SONG_RATED   = "songs.rating.created"
	SONG_CREATE_DURABLE  = "SONG_CREATE_PROCESSOR"
	SONG_RATE_DURABLE    = "SONG_RATE_PROCESSOR"
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

type UserRegistrationPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type GenreCreationPayload struct {
	GenreID   string `json:"genre_id"`
	GenreName string `json:"genre_name"`
}

type SongCreationPayload struct {
	SongID    string   `json:"song_id"`
	SongTitle string   `json:"song_title"`
	Duration  int      `json:"duration"`
	GenreIDs  []string `json:"genre_ids"`
}

type GenreSubscriptionEventPayload struct {
	GenreID string `json:"genre_id"`
	UserID  string `json:"user_id"`
}

type SongRatingPayload struct {
	SongID string `json:"song_id"`
	UserID string `json:"user_id"`
	Value  int    `json:"value"`
}
