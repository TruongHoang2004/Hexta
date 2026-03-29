package request

type PaginationRequest struct {
	Page   int32             `json:"page"`
	Limit  int32             `json:"limit"`
	Filter map[string]string `json:"filter"`
	Sort   string            `json:"sort"`
	Order  string            `json:"order"`
	Search string            `json:"search"`
}
