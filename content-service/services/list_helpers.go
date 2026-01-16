package services

import (
	"context"

	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"go.mongodb.org/mongo-driver/bson"
)

func listWithPagination[T any](
	ctx context.Context,
	p pagination.Pagination,
	filter bson.M,
	find func(context.Context, bson.M, int64, int64) ([]T, int64, error),
) (*commondtos.ItemCollectionResponse[T], error) {
	items, total, err := find(ctx, filter, p.Skip(), p.Limit())
	if err != nil {
		return nil, err
	}

	return &commondtos.ItemCollectionResponse[T]{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}
