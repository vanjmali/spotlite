package repositories

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/notifications/entities"
)

const (
	PAGE_SIZE = 10
)

var (
	ErrClosingIterator = errors.New("couldn't close iterator")
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
	qText := "INSERT INTO notifications (user_id, created_at, notification_id, notification_type, message) VALUES (?, ?, ?, ?, ?)"

	err := r.s.Query(qText,
		n.UserID,
		n.CreatedAt,
		n.NotificationID,
		n.Type,
		n.Message,
	).WithContext(ctx).Exec()

	return err
}

func (r *NotificationRepository) FindNotificationsByUserID(userID string, ctx context.Context) ([]*entities.Notification, error) {
	qText := "SELECT user_id, created_at, notification_id, notification_type, message FROM notifications WHERE user_id = ? LIMIT 10"
	iter := r.s.Query(qText, userID).WithContext(ctx).Iter()

	// allocate a slice of size PAGE_SIZE
	notifications := make([]*entities.Notification, 0, PAGE_SIZE)

	var n *entities.Notification
	for {
		n = &entities.Notification{}

		if !iter.Scan(
			&n.UserID,
			&n.CreatedAt,
			&n.NotificationID,
			&n.Type,
			&n.Message,
		) {
			break
		}
		notifications = append(notifications, n)
	}

	if err := iter.Close(); err != nil {
		return nil, ErrClosingIterator
	}

	return notifications, nil
}
