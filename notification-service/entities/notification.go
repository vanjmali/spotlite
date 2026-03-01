package entities

import (
	"time"

	"github.com/gocql/gocql"
)

type NotificationType string

const (
	NotificationNewArtist NotificationType = "ARTIST"
	NotificationNewAlbum  NotificationType = "ALBUM"
	NotificationNewSong   NotificationType = "SONG"
)

type Notification struct {
	UserID string `db:"user_id" json:"user_id"`

	// instead of using TimeUUID which combines CreatedAt and NotificationID
	// we keep them separated for easier calculations. If we used TimeUUID
	// when only the time is needed, we would have to extract it from the TimeUUID
	CreatedAt      time.Time        `db:"created_at" json:"created_at"`
	NotificationID gocql.UUID       `db:"notification_id" json:"notification_id"`
	Type           NotificationType `db:"notification_type" json:"notification_type"`
	EntityID       string           `db:"entity_id" json:"entity_id"`
	EntityName     string           `db:"entity_name" json:"entity_name"`
}
