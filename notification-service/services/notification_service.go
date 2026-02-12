package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/notification-service/entities"
	"github.com/vanjmali/spotlite/notification-service/infrastructure"
	"github.com/vanjmali/spotlite/notification-service/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrMissingUserID = errors.New("UserID is missing")

type NotificationService struct {
	r *repositories.NotificationRepository
	b *infrastructure.Broker

	tr trace.Tracer
}

func NewNotificationService(r *repositories.NotificationRepository, b *infrastructure.Broker) *NotificationService {
	tr := otel.Tracer("notification-service/notification-service")
	s := NotificationService{
		r:  r,
		b:  b,
		tr: tr,
	}
	return &s
}

// CreateNotification function is implemented only for demonstration purposes.
func (s *NotificationService) CreateNotification(np events.SubscribersBatchEventPayload, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "notification.create_notification")
	defer span.End()

	var notifType entities.NotificationType

	for _, sID := range np.SubscriberIDs {

		switch events.EntityType(np.EntityType) {
		case events.AlbumType:
			notifType = entities.NotificationNewAlbum
		case events.ArtistType:
			notifType = entities.NotificationNewArtist
		default:
			// TODO: create custom error
			return errors.New("er cn")
		}

		notificationID := gocql.TimeUUID()
		createdAt := np.CreatedAt

		n := entities.Notification{
			UserID:         sID,
			CreatedAt:      createdAt,
			Type:           notifType,
			NotificationID: notificationID,
			EntityID:       np.EntityID,
			EntityName:     np.EntityName,
		}

		if err := s.r.InsertNotification(&n, ctx); err != nil {
			span.RecordError(err)
			return err
		}

		np, err := json.Marshal(n)
		if err != nil {
			return err
		}

		s.b.Broadcast <- infrastructure.NewNotification(sID, np)
	}

	return nil
}

// FindNotifications.
func (s *NotificationService) FindInboxByUserID(ctx context.Context) ([]*entities.Notification, error) {
	ctx, span := s.tr.Start(ctx, "notification.find_notifications_by_user_id")
	defer span.End()

	userID := middlewares.GetUserIdFromContext(ctx)

	if userID == "" {
		return []*entities.Notification{}, ErrMissingUserID
	}

	ns, err := s.r.FindNotificationsByUserID(userID, ctx)
	if err != nil {
		return []*entities.Notification{}, err
	}

	return ns, nil
}
