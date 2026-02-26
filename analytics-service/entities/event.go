package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event Sourcing Model
//
// This file defines the Event type which forms the foundation of the Event Sourcing pattern.
// All user activities (song plays, subscriptions, ratings) are captured as immutable events
// stored in the events collection. The event log serves as the single source of truth for
// analytics data.
//
// Event Sourcing Benefits:
// - Immutable audit trail: complete history of all user actions
// - Temporal queries: answer "what was the state at time X?"
// - Event replay: rebuild read models or fix bugs by replaying events
// - Compliance: regulatory requirements for activity tracking
// - Performance: efficient sequential writes to event log
//
// See README.md Event Sourcing + CQRS section for architecture details.

// EventType constants represent the types of events that can occur in the system.
// These are the fundamental activities tracked for analytics.
const (
	// SongPlayedEvent occurs when a user listens to a song
	// Data: { songID, artistID, albumID, genreID, durationMS, playedAt }
	EventTypeSongPlayed = "song_played"

	// RatingCreatedEvent occurs when a user rates a song for the first time
	// Data: { songID, rating (1-5), createdAt }
	EventTypeRatingCreated = "rating_created"

	// RatingUpdatedEvent occurs when a user changes their rating for a song
	// Data: { songID, oldRating, newRating, updatedAt }
	EventTypeRatingUpdated = "rating_updated"

	// RatingDeletedEvent occurs when a user removes their rating for a song
	// Data: { songID, deletedRating, deletedAt }
	EventTypeRatingDeleted = "rating_deleted"

	// SubscriptionCreatedEvent occurs when a user subscribes to an artist or genre
	// Data: { subscriptionType (artist|genre), targetID, createdAt }
	EventTypeSubscriptionCreated = "subscription_created"

	// SubscriptionDeletedEvent occurs when a user unsubscribes from an artist or genre
	// Data: { subscriptionType (artist|genre), targetID, deletedAt }
	EventTypeSubscriptionDeleted = "subscription_deleted"

	// SongDeletedEvent occurs when a song is deleted from the system
	// This event represents cleanup for analytics (affects historical data)
	// Data: { songID, deletedAt }
	EventTypeSongDeleted = "song_deleted"
)

// AggregateType constants represent the types of entities that events relate to
const (
	AggregateTypeSong         = "song"
	AggregateTypeRating       = "rating"
	AggregateTypeSubscription = "subscription"
	AggregateTypeUser         = "user"
)

// SubscriptionType represents the granularity of a subscription event
// Users can subscribe to specific artists or entire genres
type SubscriptionType string

const (
	SubscriptionTypeArtist SubscriptionType = "artist"
	SubscriptionTypeGenre  SubscriptionType = "genre"
)

// Event represents an immutable event in the event sourcing log.
// Events are the primary persistence mechanism for the analytics service.
// Each event is written once and never modified, forming an append-only log.
//
// The Event struct is stored in MongoDB's events collection and serves as the
// single source of truth for all user activity. Read models are built by
// projecting events from this log.
type Event struct {
	// ID is the MongoDB document ID for this event
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// UserID is the ID of the user who triggered this event
	UserID string `bson:"user_id" json:"user_id" validate:"required"`

	// AggregateID is the ID of the entity this event relates to
	// For SongPlayedEvent: the song ID
	// For RatingCreatedEvent: the song ID being rated
	// For SubscriptionCreatedEvent: the artist or genre ID
	AggregateID string `bson:"aggregate_id" json:"aggregate_id" validate:"required"`

	// AggregateType indicates what kind of entity this event relates to
	// Values: "song", "rating", "subscription", "user"
	AggregateType string `bson:"aggregate_type" json:"aggregate_type" validate:"required,oneof=song rating subscription user"`

	// EventType is the specific type of event that occurred
	// Values: song_played, rating_created, rating_updated, rating_deleted,
	//         subscription_created, subscription_deleted, song_deleted
	EventType string `bson:"event_type" json:"event_type" validate:"required"`

	// Data contains the event-specific payload as key-value pairs
	// Structure depends on EventType:
	// - SongPlayedEvent: {songID, artistID, albumID, genreID, durationMS, playedAt}
	// - RatingCreatedEvent: {songID, rating, createdAt}
	// - RatingUpdatedEvent: {songID, oldRating, newRating, updatedAt}
	// - RatingDeletedEvent: {songID, deletedRating, deletedAt}
	// - SubscriptionCreatedEvent: {subscriptionType, targetID, createdAt}
	// - SubscriptionDeletedEvent: {subscriptionType, targetID, deletedAt}
	// - SongDeletedEvent: {songID, deletedAt}
	Data map[string]interface{} `bson:"data" json:"data"`

	// Version is a sequential counter for ordering events from the same aggregate
	// Starts at 1 for the first event of an aggregate and increments
	// Used together with Timestamp for deterministic ordering
	Version int64 `bson:"version" json:"version" validate:"required,min=1"`

	// Timestamp is when the event occurred in UTC
	// Used for temporal queries and sorting events
	Timestamp time.Time `bson:"timestamp" json:"timestamp" validate:"required"`

	// TraceID enables distributed tracing across service boundaries
	// Correlates this event with other operations in the same transaction
	// Passed through to NATS and downstream services for observability
	TraceID string `bson:"trace_id" json:"trace_id"`

	// SpanID enables distributed tracing within a single service
	// Identifies this specific operation in the trace context
	SpanID string `bson:"span_id" json:"span_id"`

	// Metadata contains additional context about how/where this event originated
	// Examples: {"source": "mobile-app", "ip": "192.168.1.1", "user_agent": "..."}
	// Can be used for analytics segmentation and debugging
	Metadata map[string]string `bson:"metadata" json:"metadata"`

	// CreatedAt is the server timestamp when this event was recorded
	// Set automatically by the repository layer, different from Timestamp
	// which is when the activity actually occurred
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// NewEvent creates a new Event with required fields and timestamp
// The caller must set Data, AggregateID, and AggregateType
// Version and Timestamp are managed by the repository layer
func NewEvent(userID, eventType string) *Event {
	return &Event{
		UserID:    userID,
		EventType: eventType,
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]string),
		Version:   1,
		Timestamp: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}
}

// SetData is a convenience method for setting single key-value pairs in event data
func (e *Event) SetData(key string, value interface{}) {
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
	e.Data[key] = value
}

// SetMetadata is a convenience method for setting single metadata key-value pairs
func (e *Event) SetMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// GetData retrieves a value from event data with type assertion support
// Returns nil if key doesn't exist
func (e *Event) GetData(key string) interface{} {
	if e.Data == nil {
		return nil
	}
	return e.Data[key]
}

// GetMetadata retrieves a value from metadata
// Returns empty string if key doesn't exist
func (e *Event) GetMetadata(key string) string {
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata[key]
}
