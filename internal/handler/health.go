package handler

import (
	"net/http"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Healthcheck godoc
// @Summary Health check endpoint
// @Description HealthCheck is a handler function that checks the health of the service and returns a JSON response indicating the status.
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ok",
		Message: "Service is healthy",
	}

	_ = WriteJSON(w, http.StatusOK, response)
}
