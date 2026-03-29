package dto

type UserResponse struct {
	ID          string `json:"id"`
	UserName    string `json:"user_name"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type GetUserRequest struct {
	ID string `json:"id" validate:"required"`
}

type ListUsersRequest struct {
	PaginationRequest
	Search string          `json:"search" form:"search" query:"search"`
	Filter ListUsersFilter `json:"filter"`
}

type ListUsersFilter struct {
	UserName    string `json:"user_name"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
}

type ListUsersResponse struct {
	PaginationResponse[UserResponse]
}
