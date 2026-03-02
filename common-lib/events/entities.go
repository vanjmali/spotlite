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

	SUBJECT_ENTITY_CREATED       = "content.created"
	SUBJECT_ENTITY_UPDATED       = "content.updated"
	SUBJECT_SUBSCRIBER_BATCH     = "subscribers.batch.process"
	SUBJECT_USER_CREATED         = "user.created"
	SUBJECT_GENRE_CREATED        = "genres.created"
	SUBJECT_GENRE_SUBSCRIBED     = "genres.subscription.created"
	SUBJECT_GENRE_UPDATED        = "genres.updated"
	SUBJECT_SONG_CREATED         = "songs.created"
	SUBJECT_SONG_RATED           = "songs.rating.created"
	SUBJECT_SONG_UPDATED         = "songs.updated"
	SUBJECT_SONG_DELETED         = "songs.deleted"
	SUBJECT_RATING_CREATED       = "rating.created"
	SUBJECT_RATING_UPDATED       = "rating.updated"
	SUBJECT_RATING_DELETED       = "rating.deleted"
	SUBJECT_LISTEN_CREATED       = "listen.created"
	SUBJECT_SUBSCRIPTION_CREATED = "subscription.created"
	SUBJECT_SUBSCRIPTION_DELETED = "subscription.deleted"
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
	OldRating int       `json:"old_rating"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ListenEventPayload struct {
	UserID    string    `json:"user_id"`
	SongID    string    `json:"song_id"`
	SongTitle string    `json:"song_title"`
	ArtistID  string    `json:"artist_id"`
	GenreID   string    `json:"genre_id"`
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
