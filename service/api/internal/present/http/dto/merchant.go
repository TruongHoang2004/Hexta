package dto

type CreateMerchantRequest struct {
	Name        string `json:"name" `
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type UpdateMerchantRequest struct {
	Id          string `json:"id" binding:"required"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type ListMerchantsRequest struct {
	PaginationRequest
	Search string `json:"search"`
	Filter ListMerchantFilter
}

type ListMerchantFilter struct {
	Status      string `json:"status"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type MerchantResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListMerchantsResponse struct {
	PaginationResponse[MerchantResponse]
}
