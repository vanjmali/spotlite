package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/notification-service/entities"
	"github.com/vanjmali/spotlite/notification-service/infrastructure"
	"github.com/vanjmali/spotlite/notification-service/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrMissingUserID     = errors.New("UserID is missing")
	ErrJsonMarshal       = errors.New("an error has occured while serializing json object")
	ErrInvalidEntityType = errors.New("invalid subscription entity type")
)

type NotificationService struct {
	r  *repositories.NotificationRepository
	rc *redis.Client
	b  *infrastructure.Broker

	tr trace.Tracer
}

func NewNotificationService(r *repositories.NotificationRepository, rc *redis.Client, b *infrastructure.Broker) *NotificationService {
	tr := otel.Tracer("notification-service/notification-service")
	s := NotificationService{
		r:  r,
		rc: rc,
		b:  b,
		tr: tr,
	}
	return &s
}

// CreateNotification function.
func (s *NotificationService) CreateNotification(np events.SubscribersBatchEventPayload, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "notification.create_notification")
	defer span.End()

	validRecipients, err := filterDuplicateNotifications(s.rc, np.EventID, np.SubscriberIDs, ctx)
	if err != nil {
		logging.Errorf(ctx, "failed to filter duplicate notifications: %v", err)
		return err
	}

	var notifType entities.NotificationType
	for _, sID := range validRecipients {
		switch np.EntityType {
		case events.AlbumType:
			notifType = entities.NotificationNewAlbum
		case events.ArtistType:
			notifType = entities.NotificationNewArtist
		case events.SongType:
			notifType = entities.NotificationNewSong
		default:
			logging.Warnf(ctx, "invalid entity type for notification: %s", np.EntityType)
			return ErrInvalidEntityType
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
			logging.Errorf(ctx, "failed to insert notification: %v", err)
			return err
		}

		np, err := json.Marshal(n)
		if err != nil {
			logging.Errorf(ctx, "failed to marshal notification: %v", err)
			return ErrJsonMarshal
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
		logging.Warnf(ctx, "missing user id in context while reading inbox")
		return []*entities.Notification{}, ErrMissingUserID
	}

	ns, err := s.r.FindNotificationsByUserID(userID, ctx)
	if err != nil {
		logging.Errorf(ctx, "failed to fetch notifications by user id: %v", err)
		return []*entities.Notification{}, err
	}

	return ns, nil
}

func filterDuplicateNotifications(rc *redis.Client, eventID string, userIDs []string, ctx context.Context) ([]string, error) {
	pipe := rc.Pipeline()

	cmds := make(map[string]*redis.BoolCmd)
	seen := make(map[string]struct{}, len(userIDs))
	for _, uid := range userIDs {
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}

		key := fmt.Sprintf("notif:%s:%s", eventID, uid)
		cmds[uid] = pipe.SetNX(ctx, key, "1", 15*time.Minute)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		logging.Errorf(ctx, "redis pipeline exec failed while deduplicating notifications: %v", err)
		return nil, err
	}

	var usersToNotify []string
	for uid, cmd := range cmds {
		if cmd.Val() {
			usersToNotify = append(usersToNotify, uid)
		}
	}

	return usersToNotify, nil
}
