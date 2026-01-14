package repositories

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/notifications/entities"
)

type NotificationRepository struct {
	s *gocql.Session
}

func NewNotificationRepository(s *gocql.Session) *NotificationRepository {
	r := NotificationRepository{
		s: s,
	}
	return &r
}

func (r *NotificationRepository) InsertNotification(n *entities.Notification, ctx context.Context) error {
	err := r.s.Query("INSERT INTO notifications (user_id, created_at, notification_id, type, is_read, read_at, message) VALUES (?, ?, ?, ?, ?, ?, ?)",
		n.UserID,
		n.CreatedAt,
		n.NotificationID,
		n.Type,
		n.IsRead,
		n.ReadAt,
		n.Message,
	).WithContext(ctx).Exec()

	return err
}
