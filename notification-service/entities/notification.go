package entities

import (
	"time"

	"github.com/gocql/gocql"
)

type NotificationType string

const (
	NotificationTypeNewArtist NotificationType = "new_artist"
	NotificationNewAlbum      NotificationType = "new_album"
	NotificationNewSong       NotificationType = "new_song"
)

type Notification struct {
	UserID string `db:"user_id" json:"user_id"`

	// instead of using TimeUUID which combines CreatedAt and NotificationID
	// we keep them separated for easier calculations. If we used TimeUUID
	// when only the time is needed, we would have to extract it from the TimeUUID
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	NotificationID gocql.UUID `db:"notification_id" json:"notification_id"`

	Type   NotificationType `db:"type" json:"type"`
	IsRead bool             `db:"is_read" json:"is_read"`

	// ReadAt is a pointer because the value can be nil if the notifications isn't
	// marked as read
	ReadAt  *time.Time `db:"read_at" json:"read_at"`
	Message string     `db:"message" json:"message"`
}
