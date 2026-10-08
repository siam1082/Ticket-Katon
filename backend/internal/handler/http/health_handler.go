package http

import (
	"encoding/json"
	"net/http"

	"ticket-katon-backend/internal/service"
)

type HealthHandler struct {
	svc *service.HealthService
}

func NewHealthHandler(svc *service.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := h.svc.Check(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if status["status"] != "UP" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(status)
}