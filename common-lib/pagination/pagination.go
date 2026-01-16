package pagination

import (
	"net/url"
	"strconv"
)

const (
	pageLimit   = 1_000_000
	maxPageSize = 50
)

type Pagination struct {
	Page int
	Size int
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

// NewPagination creates a Pagination instance with clamped page and size values.
func NewPagination(page, size int) Pagination {
	return Pagination{
		Page: clampInt(page, 1, pageLimit),
		Size: clampInt(size, 1, maxPageSize),
	}
}

// ParsePagination reads page/size from query params with defaults and clamps them.
func ParsePagination(q url.Values) Pagination {
	page, err := strconv.Atoi(q.Get("page"))
	if err != nil {
		page = 1
	}

	size, err := strconv.Atoi(q.Get("size"))
	if err != nil {
		size = 10
	}

	return NewPagination(page, size)
}

func (p Pagination) Skip() int64 {
	return int64((p.Page - 1) * p.Size)
}

func (p Pagination) Limit() int64 {
	return int64(p.Size)
}
