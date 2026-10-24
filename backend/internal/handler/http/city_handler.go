package http

import (
	"net/http"
	"strconv"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/service"
)

// CityHandler serves the public city autocomplete.
type CityHandler struct {
	svc *service.CityService
}

func NewCityHandler(svc *service.CityService) *CityHandler {
	return &CityHandler{svc: svc}
}

// Suggest handles GET /api/v1/cities?q=dha&limit=8
// It understands other spellings (Chittagong for Chattogram) and small typos (Rajshai for Rajshahi).
func (h *CityHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	limit := 8
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 20 {
		limit = v
	}

	q := r.URL.Query().Get("q")
	if len(q) > 60 {
		q = q[:60]
	}

	suggestions, err := h.svc.Suggest(r.Context(), q, limit)
	if err != nil {
		bookingWriteError(w, http.StatusInternalServerError, "could not load cities")
		return
	}
	if suggestions == nil {
		suggestions = []domain.CitySuggestion{}
	}
	bookingWriteJSON(w, http.StatusOK, suggestions)
}
