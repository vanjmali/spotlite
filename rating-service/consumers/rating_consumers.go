package consumers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/rating-service/services"
)

type RatingConsumer struct {
	rs *services.RatingService
}

func NewRatingConsumer(rs *services.RatingService) *RatingConsumer {
	return &RatingConsumer{
		rs: rs,
	}
}

func (c *RatingConsumer) HandleSongDelete(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongDeletePayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongDeletePayload: %v", err)
		return nil
	}

	if err := c.rs.DeleteSongRatings(ctx, p.SongID); err != nil {
		logging.Errorf(ctx, "error: an error has occured while handling delete song event: %v", err)
		return err
	}

	return nil
}
