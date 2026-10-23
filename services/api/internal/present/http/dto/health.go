package dto

type HealthResponse struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details"`
}
