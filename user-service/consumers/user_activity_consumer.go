package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/services"
)

type UserActivityConsumer struct {
	s *services.UserActivityService
}

func NewUserActivityConsumer(s *services.UserActivityService) *UserActivityConsumer {
	return &UserActivityConsumer{s: s}
}

func (c *UserActivityConsumer) HandleListenCreated(ctx context.Context, msg jetstream.Msg) error {
	var payload events.ListenEventPayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		logging.Errorf(ctx, "failed to unmarshal listen event: %v", err)
		return err
	}

	return c.s.Append(ctx, entities.UserActivity{
		EventID:    payload.EventID,
		UserID:     payload.UserID,
		Type:       entities.ActivityListen,
		EntityID:   payload.SongID,
		EntityName: payload.SongTitle,
		EntityType: "SONG",
		OccurredAt: atOrNow(payload.CreatedAt),
	})
}

func (c *UserActivityConsumer) HandleRatingCreated(ctx context.Context, msg jetstream.Msg) error {
	return c.handleRatingEvent(ctx, msg, entities.ActivityRatingCreate)
}

func (c *UserActivityConsumer) HandleRatingUpdated(ctx context.Context, msg jetstream.Msg) error {
	return c.handleRatingEvent(ctx, msg, entities.ActivityRatingUpdate)
}

func (c *UserActivityConsumer) HandleSubscriptionCreated(ctx context.Context, msg jetstream.Msg) error {
	return c.handleSubscriptionEvent(ctx, msg, entities.ActivitySubscriptionCreate)
}

func (c *UserActivityConsumer) HandleSubscriptionDeleted(ctx context.Context, msg jetstream.Msg) error {
	return c.handleSubscriptionEvent(ctx, msg, entities.ActivitySubscriptionDelete)
}

func (c *UserActivityConsumer) handleRatingEvent(
	ctx context.Context,
	msg jetstream.Msg,
	activityType entities.ActivityType,
) error {
	var payload events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		logging.Errorf(ctx, "failed to unmarshal rating event: %v", err)
		return err
	}

	ratingValue := payload.Rating
	return c.s.Append(ctx, entities.UserActivity{
		EventID:     payload.EventID,
		UserID:      payload.UserID,
		Type:        activityType,
		EntityID:    payload.SongID,
		EntityName:  payload.SongTitle,
		EntityType:  "SONG",
		RatingValue: &ratingValue,
		OccurredAt:  atOrNow(payload.CreatedAt),
	})
}

func (c *UserActivityConsumer) handleSubscriptionEvent(
	ctx context.Context,
	msg jetstream.Msg,
	activityType entities.ActivityType,
) error {
	var payload events.SubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		logging.Errorf(ctx, "failed to unmarshal subscription event: %v", err)
		return err
	}

	return c.s.Append(ctx, entities.UserActivity{
		EventID:    payload.EventID,
		UserID:     payload.UserID,
		Type:       activityType,
		EntityID:   payload.EntityID,
		EntityName: payload.EntityName,
		EntityType: string(payload.EntityType),
		OccurredAt: atOrNow(payload.CreatedAt),
	})
}

func atOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}
