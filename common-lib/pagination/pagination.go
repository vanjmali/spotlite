package pagination

const (
	pageLimit   = 1_000_000
	maxPageSize = 50
)

type Pagination struct {
	Page int
	Size int
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func NewPagination(page, size int) Pagination {
	return Pagination{
		Page: clampInt(page, 1, pageLimit),
		Size: clampInt(size, 1, maxPageSize),
	}
}

func (p Pagination) Skip() int64 {
	return int64((p.Page - 1) * p.Size)
}

func (p Pagination) Limit() int64 {
	return int64(p.Size)
}
