package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/analytics-service/entities"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/subscription"
)

// AnalyticsService defines the interface for analytics business logic operations
type AnalyticsService interface {
	StoreEvent(ctx context.Context, event *entities.Event) error
	ProjectSongPlayedEvent(ctx context.Context, userID, genreID, artistID string, timestamp time.Time) error
	ProjectRatingEvent(ctx context.Context, userID string, eventType string, rating int, oldRating int, timestamp time.Time) error
	ProjectSubscriptionEvent(ctx context.Context, userID string, eventType string, subscriptionType subscription.SubscriptionType, timestamp time.Time) error
}

// AnalyticsConsumer handles events from NATS JetStream and projects them into read models
type AnalyticsConsumer struct {
	analyticsService AnalyticsService
}

// NewConsumer creates a new AnalyticsConsumer instance with the provided analytics service
func NewConsumer(analyticsService AnalyticsService) *AnalyticsConsumer {
	return &AnalyticsConsumer{
		analyticsService: analyticsService,
	}
}

// HandleSongPlayed processes song played events - increments play count and updates genre/artist stats
func (h *AnalyticsConsumer) HandleSongPlayed(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongPlayedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SongPlayedEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSongPlayed,
		Data: map[string]interface{}{
			"songID":     p.SongID,
			"artistID":   p.ArtistID,
			"albumID":    p.AlbumID,
			"genreID":    p.GenreID,
			"durationMS": p.DurationMS,
			"playedAt":   p.PlayedAt,
		},
		Timestamp: p.PlayedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store song played event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectSongPlayedEvent(
		ctx,
		p.UserID,
		p.GenreID,
		p.ArtistID,
		p.PlayedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project song played event: %v", err)
		return err
	}

	return nil
}

// HandleRatingCreated processes rating created events - stores rating and updates average
func (h *AnalyticsConsumer) HandleRatingCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingCreatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingCreatedEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeRatingCreated,
		Data: map[string]interface{}{
			"songID":    p.SongID,
			"rating":    p.Rating,
			"createdAt": p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store rating created event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectRatingEvent(
		ctx,
		p.UserID,
		entities.EventTypeRatingCreated,
		p.Rating,
		0, // no old rating for created
		p.CreatedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project rating created event: %v", err)
		return err
	}

	return nil
}

// HandleRatingUpdated processes rating updated events - updates rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingUpdated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingUpdatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingUpdatedEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeRatingUpdated,
		Data: map[string]interface{}{
			"songID":    p.SongID,
			"oldRating": p.OldRating,
			"newRating": p.NewRating,
			"updatedAt": p.UpdatedAt,
		},
		Timestamp: p.UpdatedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store rating updated event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectRatingEvent(
		ctx,
		p.UserID,
		entities.EventTypeRatingUpdated,
		p.NewRating,
		p.OldRating,
		p.UpdatedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project rating updated event: %v", err)
		return err
	}

	return nil
}

// HandleRatingDeleted processes rating deleted events - removes rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingDeletedEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeRatingDeleted,
		Data: map[string]interface{}{
			"songID":        p.SongID,
			"deletedRating": p.DeletedRating,
			"deletedAt":     p.DeletedAt,
		},
		Timestamp: p.DeletedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store rating deleted event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectRatingEvent(
		ctx,
		p.UserID,
		entities.EventTypeRatingDeleted,
		p.DeletedRating,
		0, // no old rating for deleted
		p.DeletedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project rating deleted event: %v", err)
		return err
	}

	return nil
}

// HandleSubscriptionCreated processes subscription created events - increments subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionCreatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionCreatedEventPayload: %v", err)
		return nil
	}

	// Convert string from event payload to type-safe enum at boundary
	var subType subscription.SubscriptionType
	switch p.SubscriptionType {
	case string(subscription.ArtistSubscription):
		subType = subscription.ArtistSubscription
	case string(subscription.GenreSubscription):
		subType = subscription.GenreSubscription
	default:
		logging.Errorf(ctx, "critical: invalid subscription type: %s", p.SubscriptionType)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSubscriptionCreated,
		Data: map[string]interface{}{
			"subscriptionType": p.SubscriptionType,
			"targetID":         p.TargetID,
			"createdAt":        p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store subscription created event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectSubscriptionEvent(
		ctx,
		p.UserID,
		entities.EventTypeSubscriptionCreated,
		subType,
		p.CreatedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project subscription created event: %v", err)
		return err
	}

	return nil
}

// HandleSubscriptionDeleted processes subscription deleted events - decrements subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionDeletedEventPayload: %v", err)
		return nil
	}

	// Convert string from event payload to type-safe enum at boundary
	var subType subscription.SubscriptionType
	switch p.SubscriptionType {
	case string(subscription.ArtistSubscription):
		subType = subscription.ArtistSubscription
	case string(subscription.GenreSubscription):
		subType = subscription.GenreSubscription
	default:
		logging.Errorf(ctx, "critical: invalid subscription type: %s", p.SubscriptionType)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSubscriptionDeleted,
		Data: map[string]interface{}{
			"subscriptionType": p.SubscriptionType,
			"targetID":         p.TargetID,
			"deletedAt":        p.DeletedAt,
		},
		Timestamp: p.DeletedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store subscription deleted event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectSubscriptionEvent(
		ctx,
		p.UserID,
		entities.EventTypeSubscriptionDeleted,
		subType,
		p.DeletedAt,
	); err != nil {
		logging.Errorf(ctx, "failed to project subscription deleted event: %v", err)
		return err
	}

	return nil
}

// HandleSongDeleted processes song deleted events - stores event for audit trail
func (h *AnalyticsConsumer) HandleSongDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongDeletedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SongDeletedEventPayload: %v", err)
		return nil
	}

	// Store event in event store for audit trail
	event := &entities.Event{
		UserID:    "", // system event, no specific user
		EventType: entities.EventTypeSongDeleted,
		Data: map[string]interface{}{
			"songID":    p.SongID,
			"deletedAt": p.DeletedAt,
		},
		Timestamp: p.DeletedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store song deleted event: %v", err)
		return err
	}

	// Note: Read model cleanup/archiving can be implemented in future if needed
	// Currently just storing the event for audit trail
	return nil
}
