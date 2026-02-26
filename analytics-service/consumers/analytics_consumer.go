package consumers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
)

// AnalyticsConsumer handles events from NATS JetStream and projects them into read models
type AnalyticsConsumer struct {
	// TODO: Add service dependencies (EventStoreService, ReadModelService)
}

// NewConsumer creates a new AnalyticsConsumer instance
func NewConsumer() *AnalyticsConsumer {
	return &AnalyticsConsumer{}
}

// HandleSongPlayed processes song played events - increments play count and updates genre/artist stats
func (h *AnalyticsConsumer) HandleSongPlayed(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongPlayedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SongPlayedEventPayload: %v", err)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Update user_analytics read model (TotalSongsPlayed, SongsByGenre, TopArtists)
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleRatingCreated processes rating created events - stores rating and updates average
func (h *AnalyticsConsumer) HandleRatingCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingCreatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingCreatedEventPayload: %v", err)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Update user_analytics read model (AverageRating)
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleRatingUpdated processes rating updated events - updates rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingUpdated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingUpdatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingUpdatedEventPayload: %v", err)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Recalculate user_analytics read model (AverageRating)
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleRatingDeleted processes rating deleted events - removes rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingDeletedEventPayload: %v", err)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Recalculate user_analytics read model (AverageRating)
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleSubscriptionCreated processes subscription created events - increments subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionCreatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionCreatedEventPayload: %v", err)
		return nil
	}

	if p.SubscriptionType != "artist" && p.SubscriptionType != "genre" {
		logging.Errorf(ctx, "critical: invalid subscription type: %s", p.SubscriptionType)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Update user_analytics read model (SubscribedArtistsCount if type is "artist")
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleSubscriptionDeleted processes subscription deleted events - decrements subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionDeletedEventPayload: %v", err)
		return nil
	}

	if p.SubscriptionType != "artist" && p.SubscriptionType != "genre" {
		logging.Errorf(ctx, "critical: invalid subscription type: %s", p.SubscriptionType)
		return nil
	}

	// TODO: Store event in event store
	// TODO: Update user_analytics read model (SubscribedArtistsCount if type is "artist")
	// TODO: Append activity to user_activity_history
	return nil
}

// HandleSongDeleted processes song deleted events - removes song from user analytics
func (h *AnalyticsConsumer) HandleSongDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SongDeletedEventPayload: %v", err)
		return nil
	}

	// TODO: Store in event store and archive from read models
	return nil
}
