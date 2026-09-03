package shared

type PaginationParams struct {
	Limit  uint `query:"limit" default:"50"`
	Offset uint `query:"offset" default:"0"`
}

type PaginatedResponse[T any] struct {
	Total  uint `json:"total"`
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
	Count  uint `json:"count"`
	Items  []T  `json:"items"`
}

func Paginate[T any](items []T, total uint, limit uint, offset uint) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Count:  uint(len(items)),
		Items:  items,
	}
}
