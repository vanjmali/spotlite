package consumers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/subscription-service/services"
)

type NatsConsumer struct {
	ss *services.SubscriptionService
}

func NewConsumer(ss *services.SubscriptionService) *NatsConsumer {
	return &NatsConsumer{
		ss: ss,
	}
}

func (h *NatsConsumer) HandleArtistCreated(ctx context.Context, msg jetstream.Msg) error {
	var payload events.ArtistEventPayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		log.Printf("CRITICAL: Failed to unmarshal ArtistCreatedPayload: %v", err)
		return nil
	}

	log.Print(payload)

	return nil
}
