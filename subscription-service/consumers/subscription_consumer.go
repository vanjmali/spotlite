package consumers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
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
		logging.Errorf(ctx, "critical: failed to unmarshal EntityCreatedPayload: %v", err)
		return nil
	}

	err := h.ss.NotifySubscribers(ctx, p)
	if err != nil {
		// if an error has occured while parsing UUID, converting from string to primitive.objectID,
		// another try won't make a difference and we want to abort
		if errors.Is(err, repositories.ErrUUIDParse) {
			return nil
		}

		// errors can be caused because of the database being down, network or any
		// other infrastructure issues so we want to retry it just in case
		return err
	}

	return nil
}

func (h *SubscriptionConsumer) HandleEntityUpdated(ctx context.Context, msg jetstream.Msg) error {
	return nil
}
