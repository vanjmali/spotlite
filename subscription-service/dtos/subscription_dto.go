package dtos

import (
	"github.com/vanjmali/spotlite/common-lib/subscription"
)

type CreateSubscriptionDto struct {
	EntityID string                        `json:"entity_id" validate:"required,validentityid"`
	Type     subscription.SubscriptionType `json:"sub_type" validate:"required,validsubtype"`
}
