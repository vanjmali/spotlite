package dtos

// ItemCollectionResponse represents a paginated collection of items.
type ItemCollectionResponse[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}
