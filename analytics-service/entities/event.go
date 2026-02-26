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

	// Timestamp is when the event occurred in UTC
	// Used for temporal queries and sorting events
	Timestamp time.Time `bson:"timestamp" json:"timestamp" validate:"required"`
}
