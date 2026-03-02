package consumers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/content/services"
)

type ContentConsumer struct {
	ss *services.SongService
}

func NewContentConsumer(ss *services.SongService) *ContentConsumer {
	return &ContentConsumer{
		ss: ss,
	}
}

func (c *ContentConsumer) HandleSongDelete(ctx context.Context, msg jetstream.Msg) error {
	var p events.SongDeletePayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "error: failed to unmarshal SongDeletePayload: %v", err)
		return nil
	}

	if err := c.ss.DeleteSong(ctx, p.SongID); err != nil {
		if errors.Is(err, services.ErrSongNotFound) {
			logging.Warnf(ctx, "song %s already deleted; acknowledging duplicate delete event", p.SongID)
			return nil
		}
		logging.Errorf(ctx, "error: an error has occured while handling delete song event: %v", err)
		return err
	}

	return nil
}
