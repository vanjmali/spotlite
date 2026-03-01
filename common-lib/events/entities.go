package events

import (
	"time"
)

type EntityType string

const (
	ArtistType EntityType = "ARTIST"
	AlbumType  EntityType = "ALBUM"
	SongType   EntityType = "SONG"

	CONTENT_STREAM       = "CONTENT"
	SUBSCRIPTIONS_STREAM = "SUBSCRIPTIONS"
	USERS_STREAM         = "USERS"
	SONGS_STREAM         = "SONGS"
	GENRES_STREAM        = "GENRES"
	ARTISTS_STREAM       = "ARTISTS"
	RATINGS_STREAM       = "RATINGS"
	LISTENS_STREAM       = "LISTENS"

	SUBJECT_ENTITY_CREATED = "content.created"
	ENTITY_CREATE_DURABLE  = "ENTITY_CREATOR"

	SUBJECT_ENTITY_UPDATED = "content.updated"
	ENTITY_UPDATE_DURABLE  = "ENTITY_UPDATER"

	SUBJECT_SUBSCRIBER_BATCH = "subscribers.batch.process"
	SUB_DURABLE              = "SUBSCRIBER_PROCESSOR"

	SUBJECT_USER_CREATED = "user.created"
	USER_DURABLE         = "USER_PROCESSOR"

	SUBJECT_GENRE_CREATED    = "genres.created"
	SUBJECT_GENRE_SUBSCRIBED = "genres.subscription.created"
	SUBJECT_GENRE_UPDATED    = "genres.updated"
	GENRE_CREATE_DURABLE     = "GENRE_CREATE_PROCESSOR"
	GENRE_SUB_DURABLE        = "GENRE_SUB_PROCESSOR"
	GENRE_UPDATE_DURABLE     = "GENRE_UPDATE_PROCESSOR"

	SUBJECT_SONG_CREATED               = "songs.created"
	SUBJECT_SONG_RATED                 = "songs.rating.created"
	SUBJECT_SONG_UPDATED               = "songs.updated"
	SUBJECT_SONG_DELETED               = "songs.deleted"
	SONG_CREATE_DURABLE                = "SONG_CREATE_PROCESSOR"
	SONG_RATE_DURABLE                  = "SONG_RATE_PROCESSOR"
	SONG_UPDATE_DURABLE                = "SONG_UPDATE_PROCESSOR"
	SONG_DELETE_DURABLE_RATING         = "SONG_DELETE_PROCESSOR_RATING"
	SONG_DELETE_DURABLE_RECOMMENDATION = "SONG_DELETE_PROCESSOR_RECOMMENDATION"
	SONG_DELETE_DURABLE_CONTENT        = "SONG_DELETE_PROCESSOR_CONTENT"

	SUBJECT_RATING_CREATED       = "rating.created"
	SUBJECT_RATING_UPDATED       = "rating.updated"
	RATING_CREATE_DURABLE        = "RATING_CREATE_PROCESSOR"
	RATING_UPDATE_DURABLE        = "RATING_UPDATE_PROCESSOR"
	RATING_CREATE_ACTION_DURABLE = "RATING_CREATE_ACTION_PROCESSOR"
	RATING_UPDATE_ACTION_DURABLE = "RATING_UPDATE_ACTION_PROCESSOR"

	SUBJECT_LISTEN_CREATED = "listen.created"
	LISTEN_DURABLE         = "LISTEN_PROCESSOR"

	SUBJECT_SUBSCRIPTION_CREATED        = "subscription.created"
	SUBJECT_SUBSCRIPTION_DELETED        = "subscription.deleted"
	SUBSCRIPTION_CREATED_ACTION_DURABLE = "SUBSCRIPTION_CREATED_ACTION_PROCESSOR"
	SUBSCRIPTION_DELETED_ACTION_DURABLE = "SUBSCRIPTION_DELETED_ACTION_PROCESSOR"
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
	SongID      string   `json:"song_id"`
	SongTitle   string   `json:"song_title"`
	Duration    int      `json:"duration"`
	GenreIDs    []string `json:"genre_ids"`
	ArtistNames []string `json:"artist_names"`
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

type SongUpdatePayload struct {
	SongID      string   `json:"song_id"`
	SongTitle   string   `json:"song_title"`
	Duration    int      `json:"duration"`
	GenreIDs    []string `json:"genre_ids"`
	ArtistNames []string `json:"artist_names"`
}

type RatingEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	SongTitle string    `json:"song_title"`
	Rating    int       `json:"rating"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ListenEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	SongTitle string    `json:"song_title"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type SubscriptionEventPayload struct {
	UserID     string                 `json:"user_id"`
	EntityID   string                 `json:"entity_id"`
	EntityName string                 `json:"entity_name"`
	EntityType SubscriptionEntityType `json:"entity_type"`
	EventID    string                 `json:"event_id"`
	CreatedAt  time.Time              `json:"created_at"`
}

type SongDeletePayload struct {
	SongID string `json:"song_id"`
}
