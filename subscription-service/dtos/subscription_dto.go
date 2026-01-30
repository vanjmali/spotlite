package dtos

import "github.com/vanjmali/spotlite/subscriptions/entities"

type CreateSubscriptionDto struct {
	EntityID string                    `json:"entity_id"`
	Type     entities.SubscriptionType `json:"sub_type"`
}
