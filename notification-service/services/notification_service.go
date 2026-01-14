package services

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/notifications/entities"
	"github.com/vanjmali/spotlite/notifications/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

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

// CreateNotification function is implemented only for demonstration purposes
func (s *NotificationService) CreateNotification(ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "notification.create_notification")
	defer span.End()

	userID := middlewares.GetUserIdFromContext(ctx)
	createdAt := time.Now()
	notificationType := entities.NotificationNewAlbum
	isRead := false
	message := "Your favorite artist Milan has added a new album called Ulica"
	notificationID := gocql.TimeUUID()

	n := entities.Notification{
		UserID:         userID,
		CreatedAt:      createdAt,
		Type:           notificationType,
		IsRead:         isRead,
		ReadAt:         nil,
		Message:        message,
		NotificationID: notificationID,
	}

	if err := s.r.InsertNotification(&n, ctx); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
