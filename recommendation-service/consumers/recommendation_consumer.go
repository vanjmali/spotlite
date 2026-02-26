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
		logging.Errorf(ctx, "critical: failed to unmarshal GenreCreationPayload: %v", err)
		return nil
	}

	if err := c.rs.CreateGenre(p, ctx); err != nil {
		return err
	}

	return nil
}
