package api

import (
	"net/http"
)

type (
	// The HealthAPI exposes HTTP endpoints for checking the health and readiness of the server.
	HealthAPI struct{}
)

// NewHealthAPI returns a new instance of the HealthAPI type.
func NewHealthAPI() *HealthAPI {
	return &HealthAPI{}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *HealthAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health", api.Health)
	mux.HandleFunc("GET /api/v1/ready", api.Ready)
}

type (
	// The HealthResponse type represents the response body returned when calling HealthAPI.Health.
	HealthResponse struct{}
)

// Health handles an inbound HTTP request to check the health of the server. On success, it responds with
// an http.StatusOK code and a JSON-encoded HealthResponse.
func (api *HealthAPI) Health(w http.ResponseWriter, r *http.Request) {
	write(w, http.StatusOK, HealthResponse{})
}

type (
	// The ReadyResponse type represents the response body returned when calling HealthAPI.Ready.
	ReadyResponse struct{}
)

// Ready handles an inbound HTTP request to check the readiness of the server. On success, it responds with
// an http.StatusOK code and a JSON-encoded ReadyResponse.
func (api *HealthAPI) Ready(w http.ResponseWriter, r *http.Request) {
	write(w, http.StatusOK, ReadyResponse{})
}
