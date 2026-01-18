package services

import (
	"context"
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/notifications/entities"
	"github.com/vanjmali/spotlite/notifications/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrMissingUserID = errors.New("UserID is missing")

type NotificationService struct {
	r  *repositories.NotificationRepository
	tr trace.Tracer
}

func NewNotificationService(r *repositories.NotificationRepository) *NotificationService {
	tr := otel.Tracer("notification-service/notification-service")
	s := NotificationService{
		r:  r,
		tr: tr,
	}
	return &s
}

// CreateNotification function is implemented only for demonstration purposes.
func (s *NotificationService) CreateNotification(ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "notification.create_notification")
	defer span.End()

	userID := middlewares.GetUserIdFromContext(ctx)
	createdAt := time.Now()
	notificationType := entities.NotificationNewAlbum
	message := "Your favorite artist Milan has added a new album called Ulica"
	notificationID := gocql.TimeUUID()

	n := entities.Notification{
		UserID:         userID,
		CreatedAt:      createdAt,
		Type:           notificationType,
		Message:        message,
		NotificationID: notificationID,
	}

	if err := s.r.InsertNotification(&n, ctx); err != nil {
		span.RecordError(err)
		return err
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
