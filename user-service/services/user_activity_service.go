package services

import (
	"context"

	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/user-service/entities"
)

type UserActivityRepository interface {
	Append(ctx context.Context, activity entities.UserActivity) error
	ListByUserID(ctx context.Context, userID string, skip int64, limit int64) ([]entities.UserActivity, int64, error)
}

type UserActivityService struct {
	r UserActivityRepository
}

func NewUserActivityService(r UserActivityRepository) *UserActivityService {
	return &UserActivityService{r: r}
}

func (s *UserActivityService) Append(ctx context.Context, activity entities.UserActivity) error {
	return s.r.Append(ctx, activity)
}

func (s *UserActivityService) ListMyActivities(
	ctx context.Context,
	page int,
	size int,
) (*commondtos.ItemCollectionResponse[entities.UserActivity], error) {
	userID := middlewares.GetUserIdFromContext(ctx)
	p := pagination.NewPagination(page, size)

	items, total, err := s.r.ListByUserID(ctx, userID, p.Skip(), p.Limit())
	if err != nil {
		return nil, err
	}

	return &commondtos.ItemCollectionResponse[entities.UserActivity]{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}
