package consumers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/notification-service/services"
)

type NotificationConsumer struct {
	ns *services.NotificationService
}

func NewConsumer(ns *services.NotificationService) *NotificationConsumer {
	return &NotificationConsumer{
		ns: ns,
	}
}

func (h *NotificationConsumer) HandleSubscribersBatch(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscribersBatchEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscribersBatchEventPayload: %v", err)
		return nil
	}

	err := h.ns.CreateNotification(p, ctx)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidEntityType), errors.Is(err, services.ErrJsonMarshal):
			return nil
		default:
			return err
		}
	}

	return nil
}
