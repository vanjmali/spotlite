package dtos

import (
	"context"
	"errors"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrObjectIdCastFailed = errors.New("failed to convert hex to objectId")

// ItemCollectionResponse represents a paginated collection of items.
type ItemCollectionResponse[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// handleListResponse is a helper function to handle list responses for different entities.
func HandleListResponse(w http.ResponseWriter, r *http.Request, logLabel string, fetch func(context.Context) (any, error)) {
	resp, err := fetch(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
		default:
			logging.Errorf(r.Context(), "failed to list %s: %v", logLabel, err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	if err := respond.OkJson(w, resp); err != nil {
		logging.Errorf(r.Context(), "failed to write list %s response: %v", logLabel, err)
	}
}

func ListWithPagination[T any](
	ctx context.Context,
	p pagination.Pagination,
	filter bson.M,
	find func(context.Context, bson.M, int64, int64) ([]T, int64, error),
) (*ItemCollectionResponse[T], error) {
	items, total, err := find(ctx, filter, p.Skip(), p.Limit())
	if err != nil {
		logging.Errorf(ctx, "list query failed: %v", err)
		return nil, err
	}

	return &ItemCollectionResponse[T]{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}
