package http


import (
	"encoding/json"
	"errors"
	"net/http"

	"ticket-katon-backend/internal/middleware"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/service"
)

type SeatLockHandler struct {
	lockService *service.SeatLockService
}

func NewSeatLockHandler(lockService *service.SeatLockService) *SeatLockHandler {
	return &SeatLockHandler{lockService: lockService}
}

type HoldSeatsRequest struct {
	TripID      string   `json:"trip_id"`
	SeatNumbers []string `json:"seat_numbers"`
}

type HoldSeatsResponse struct {
	Status           string   `json:"status"`
	Message          string   `json:"message"`
	TripID           string   `json:"trip_id"`
	SeatsHeld        []string `json:"seats_held"`
	ExpiresInSeconds int      `json:"expires_in_seconds"`
}

type ReleaseSeatsRequest struct {
	TripID      string   `json:"trip_id"`
	SeatNumbers []string `json:"seat_numbers"`
}

// Hold handles POST /api/v1/seats/hold
func (h *SeatLockHandler) Hold(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil || claims.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	userID := claims.UserID

	var req HoldSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request payload"})
		return
	}

	if req.TripID == "" || len(req.SeatNumbers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "trip_id and seat_numbers are required"})
		return
	}

	err := h.lockService.HoldSeats(r.Context(), req.TripID, req.SeatNumbers, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrSeatsAlreadyHeld) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "One or more selected seats are currently locked or unavailable",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to acquire seat hold"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HoldSeatsResponse{
		Status:           "success",
		Message:          "Seats held successfully",
		TripID:           req.TripID,
		SeatsHeld:        req.SeatNumbers,
		ExpiresInSeconds: int(service.LockTTL.Seconds()),
	})
}

// Release handles POST /api/v1/seats/release
func (h *SeatLockHandler) Release(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil || claims.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	userID := claims.UserID

	var req ReleaseSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request payload"})
		return
	}

	if req.TripID == "" || len(req.SeatNumbers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "trip_id and seat_numbers are required"})
		return
	}

	err := h.lockService.ReleaseSeats(r.Context(), req.TripID, req.SeatNumbers, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to release seats"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Seats released successfully",
	})
}