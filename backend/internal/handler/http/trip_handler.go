package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
	"ticket-katon-backend/internal/service"
)

type TripHandler struct {
	svc *service.TripService
}

func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{svc: svc}
}

func (h *TripHandler) SearchTrips(w http.ResponseWriter, r *http.Request) {
	origin := r.URL.Query().Get("origin")
	destination := r.URL.Query().Get("destination")
	travelDate := r.URL.Query().Get("date")
	transportType := r.URL.Query().Get("type")

	params := repository.TripSearchParams{
		Origin:        origin,
		Destination:   destination,
		TravelDate:    travelDate,
		TransportType: transportType,
	}

	results, err := h.svc.Search(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func (h *TripHandler) GetTripSeats(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "tripID")
	layout, err := h.svc.GetSeats(r.Context(), tripID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, `{"error":"trip not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(layout)
}