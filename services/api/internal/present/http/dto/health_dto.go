package dto

// HealthCheckResponse represents the health check status and component details.
type HealthCheckResponse struct {
	Status  string            `json:"status" example:"up"`
	Details map[string]string `json:"details"`
}
