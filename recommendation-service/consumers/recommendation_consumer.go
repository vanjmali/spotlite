package consumers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/recommendation-service/services"
)

type RecommendationConsumer struct {
	rs *services.RecommendationService
}

func NewRecommendationConsumer(rs *services.RecommendationService) *RecommendationConsumer {
	return &RecommendationConsumer{
		rs: rs,
	}
}

func (c *RecommendationConsumer) HandleUserRegistration(ctx context.Context, msg jetstream.Msg) error {
	var p events.UserRegistrationPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal UserRegistrationPayload: %v", err)
		return nil
	}

	if err := c.rs.CreateUser(p, ctx); err != nil {
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleGenreCreation(ctx context.Context, msg jetstream.Msg) error {
	var p events.GenreCreationPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal GenreCreationPayload: %v", err)
		return nil
	}

	if err := c.rs.CreateGenre(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling genre creation event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleSongCreation(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongCreationPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongCreationPayload: %v", err)
		return nil
	}

	if err := c.rs.CreateSong(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling song creation event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleGenreSubscription(ctx context.Context, msg jetstream.Msg) error {
	var p events.GenreSubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal GenreSubscriptionEventPayload: %v", err)
		return nil
	}

	logging.Infof(ctx, "%s", p)

	if err := c.rs.CreateSubscription(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling genre subscription event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleSongRating(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongRatingPayload: %v", err)
		return nil
	}

	if err := c.rs.CreateRating(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling song rating event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleGenreUpdate(ctx context.Context, msg jetstream.Msg) error {
	var p events.EntityUpdatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal EntityUpdatedEventPayload: %v", err)
		return nil
	}

	if err := c.rs.UpdateGenre(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling genre update event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleSongUpdate(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongUpdatePayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongUpdatePayload: %v", err)
		return nil
	}

	if err := c.rs.UpdateSong(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling song update event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleSongDelete(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongDeletePayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongDeletePayload: %v", err)
		return nil
	}

	if err := c.rs.DeleteSong(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling song delete event: %v", err)
		return err
	}

	return nil
}

func (c *RecommendationConsumer) HandleRatingUpdate(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	logging.Infof(ctx, "%s", p)

	if err := c.rs.UpdateRating(p, ctx); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling rating update event: %v", err)
		return err
	}

	return nil
}
