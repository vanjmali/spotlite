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

	Type    NotificationType `db:"notification_type" json:"notification_type"`
	Message string           `db:"message" json:"message"`
}

type NotificationEvent struct {
	UserID         string     `json:"user_id"`
	CreatedAt      time.Time  `json:"created_at"`
	NotificationID gocql.UUID `json:"notification_id"`
	Message        string     `json:"message"`
}
