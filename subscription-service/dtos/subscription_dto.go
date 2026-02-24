package dtos

import (
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/entities"
)

type CreateSubscriptionDto struct {
	EntityID string                        `json:"entity_id" validate:"required,validentityid"`
	Type     subscription.SubscriptionType `json:"sub_type" validate:"required,validsubtype"`
}

type EntitySubscriberCountDTO struct {
	SubscriberCount int64 `json:"sub_count"`
}

type SubsListResponseDto = commondtos.ItemCollectionResponse[entities.Subscription]
