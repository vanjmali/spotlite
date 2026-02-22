package events

import (
	"time"
)

type EntityType string

const (
	ArtistType EntityType = "ARTIST"
	AlbumType  EntityType = "ALBUM"

	CONTENT_STREAM       = "CONTENT"
	SUBSCRIPTIONS_STREAM = "SUBSCRIPTIONS"

	SUBJECT_ENTITY_CREATED = "content.created"
	ENTITY_CREATE_DURABLE  = "ENTITY_CREATOR"

	SUBJECT_ENTITY_UPDATED = "content.updated"
	ENTITY_UPDATE_DURABLE  = "ENTITY_UPDATER"

	SUBJECT_SUBSCRIBER_BATCH = "subscribers.batch.process"
	SUB_DURABLE              = "SUBSCRIBER_PROCESSOR"
)

type EntityCreatedEventPayload struct {
	TargetIDs  []string   `json:"target_ids"`
	CreatedAt  time.Time  `json:"created_at"`
	EntityID   string     `json:"entity_id"`
	EntityName string     `json:"entity_name"`
	EntityType EntityType `json:"entity_type"`
	EventID    string     `json:"event_id"`
}

type EntityUpdatedEventPayload struct {
	EntityID   string `json:"entity_id"`
	EntityName string `json:"entity_name"`
}

type SubscribersBatchEventPayload struct {
	EntityID      string     `json:"entity_id"`
	EntityName    string     `json:"entity_name"`
	EntityType    EntityType `json:"entity_type"`
	CreatedAt     time.Time  `json:"created_at"`
	SubscriberIDs []string   `json:"subscriber_ids"`
	EventID       string     `json:"event_id"`
}
