package consumers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/subscription-service/services"
)

type SubscriptionConsumer struct {
	ss *services.SubscriptionService
}

func NewConsumer(ss *services.SubscriptionService) *SubscriptionConsumer {
	return &SubscriptionConsumer{
		ss: ss,
	}
}

func (h *SubscriptionConsumer) HandleEntityCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.EntityCreatedEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		log.Printf("CRITICAL: Failed to unmarshal EntityCreatedPayload: %v", err)
		msg.Ack()
		return nil
	}

	err := h.ss.NotifySubscribers(ctx, p)
	if err != nil {
		msg.Nak()
		return err
	}

	msg.Ack()
	return nil
}
