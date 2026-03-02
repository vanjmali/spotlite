package consumers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/analytics-service/entities"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/subscription"
)

// AnalyticsService defines the interface for analytics business logic operations
type AnalyticsService interface {
	StoreEvent(ctx context.Context, event *entities.Event) error
	ProjectSongPlayedEvent(ctx context.Context, userID, genreID, artistID string) error
	ProjectRatingEvent(ctx context.Context, userID string, eventType string, rating int, oldRating int) error
	ProjectSubscriptionEvent(ctx context.Context, userID string, eventType string, subscriptionType subscription.SubscriptionType) error
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

// HandleListenCreated processes listen created events - updates analytics and appends activity
func (h *AnalyticsConsumer) HandleListenCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.ListenEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal ListenEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSongPlayed,
		Data: map[string]interface{}{
			"songID":    p.SongID,
			"artistID":  p.ArtistID,
			"genreID":   p.GenreID,
			"createdAt": p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
	}

	if err := h.analyticsService.StoreEvent(ctx, event); err != nil {
		logging.Errorf(ctx, "failed to store listen event: %v", err)
		return err
	}

	// Project event to read models
	if err := h.analyticsService.ProjectSongPlayedEvent(
		ctx,
		p.UserID,
		p.GenreID,
		p.ArtistID,
	); err != nil {
		logging.Errorf(ctx, "failed to project listen event: %v", err)
		return err
	}

	return nil
}

// HandleRatingCreated processes rating created events - stores rating and updates average
func (h *AnalyticsConsumer) HandleRatingCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
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
	); err != nil {
		logging.Errorf(ctx, "failed to project rating created event: %v", err)
		return err
	}

	return nil
}

// HandleRatingUpdated processes rating updated events - updates rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingUpdated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeRatingUpdated,
		Data: map[string]interface{}{
			"songID":    p.SongID,
			"newRating": p.Rating,
			"updatedAt": p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
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
		p.Rating,
		p.OldRating,
	); err != nil {
		logging.Errorf(ctx, "failed to project rating updated event: %v", err)
		return err
	}

	return nil
}

// HandleRatingDeleted processes rating deleted events - removes rating and recalculates average
func (h *AnalyticsConsumer) HandleRatingDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeRatingDeleted,
		Data: map[string]interface{}{
			"songID":        p.SongID,
			"deletedRating": p.Rating,
			"deletedAt":     p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
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
		p.Rating,
		0, // no old rating for deleted
	); err != nil {
		logging.Errorf(ctx, "failed to project rating deleted event: %v", err)
		return err
	}

	return nil
}

// HandleSubscriptionCreated processes subscription created events - increments subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionEventPayload: %v", err)
		return nil
	}

	// Convert entity type to subscription type
	var subType subscription.SubscriptionType
	switch p.EntityType {
	case events.SubscriptionEntityArtist:
		subType = subscription.ArtistSubscription
	case events.SubscriptionEntityGenre:
		subType = subscription.GenreSubscription
	default:
		logging.Errorf(ctx, "critical: invalid entity type: %s", p.EntityType)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSubscriptionCreated,
		Data: map[string]interface{}{
			"entityType": string(p.EntityType),
			"entityID":   p.EntityID,
			"createdAt":  p.CreatedAt,
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
	); err != nil {
		logging.Errorf(ctx, "failed to project subscription created event: %v", err)
		return err
	}

	return nil
}

// HandleSubscriptionDeleted processes subscription deleted events - decrements subscribed count
func (h *AnalyticsConsumer) HandleSubscriptionDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionEventPayload: %v", err)
		return nil
	}

	// Convert entity type to subscription type
	var subType subscription.SubscriptionType
	switch p.EntityType {
	case events.SubscriptionEntityArtist:
		subType = subscription.ArtistSubscription
	case events.SubscriptionEntityGenre:
		subType = subscription.GenreSubscription
	default:
		logging.Errorf(ctx, "critical: invalid entity type: %s", p.EntityType)
		return nil
	}

	// Store event in event store
	event := &entities.Event{
		UserID:    p.UserID,
		EventType: entities.EventTypeSubscriptionDeleted,
		Data: map[string]interface{}{
			"entityType": string(p.EntityType),
			"entityID":   p.EntityID,
			"deletedAt":  p.CreatedAt,
		},
		Timestamp: p.CreatedAt,
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
	); err != nil {
		logging.Errorf(ctx, "failed to project subscription deleted event: %v", err)
		return err
	}

	return nil
}
