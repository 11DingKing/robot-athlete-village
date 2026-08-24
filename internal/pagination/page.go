package pagination

type Request struct {
	Limit, Offset int
	Sort          string
}

func Normalize(limit, offset int) Request {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return Request{Limit: limit, Offset: offset, Sort: "created_at DESC"}
}

type Result[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}
