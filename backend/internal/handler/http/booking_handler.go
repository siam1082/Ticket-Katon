package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/middleware"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/service"
)

type BookingHandler struct {
	svc service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func bookingWriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func bookingWriteError(w http.ResponseWriter, status int, msg string) {
	bookingWriteJSON(w, status, map[string]string{"error": msg})
}

func bookingClaims(r *http.Request) (*token.CustomClaims, bool) {
	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		return nil, false
	}
	return claims, true
}

// bookingHandleError maps service errors to HTTP responses.
func bookingHandleError(w http.ResponseWriter, tag string, err error) {
	msg := err.Error()
	var be *domain.BookingError
	if errors.As(err, &be) {
		msg = be.Msg
	}

	switch {
	case errors.Is(err, domain.ErrBookingInvalid):
		bookingWriteError(w, http.StatusBadRequest, msg)
	case errors.Is(err, domain.ErrHoldInvalid),
		errors.Is(err, domain.ErrSeatsTaken),
		errors.Is(err, domain.ErrPaymentDuplicate):
		bookingWriteError(w, http.StatusConflict, msg)
	case errors.Is(err, domain.ErrBookingNotFound):
		bookingWriteError(w, http.StatusNotFound, msg)
	default:
		log.Printf("[%s ERROR] %+v", tag, err)
		bookingWriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ---------------------------------------------------------------------------
// handlers
// ---------------------------------------------------------------------------

// Confirm handles POST /api/v1/bookings/confirm
func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	var req domain.ConfirmBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	resp, err := h.svc.ConfirmBooking(r.Context(), claims.UserID, &req)
	if err != nil {
		bookingHandleError(w, "CONFIRM BOOKING", err)
		return
	}

	bookingWriteJSON(w, http.StatusCreated, resp)
}

// GetByPNR handles GET /api/v1/bookings/{pnr}
// Only the booking owner or staff (operator admin / super admin) can view it.
func (h *BookingHandler) GetByPNR(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pnr := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "pnr")))
	if pnr == "" {
		bookingWriteError(w, http.StatusBadRequest, "pnr is required")
		return
	}

	booking, tickets, err := h.svc.GetBookingByPNR(r.Context(), pnr)
	if err != nil {
		bookingHandleError(w, "GET BOOKING", err)
		return
	}

	isStaff := claims.Role == domain.RoleOperatorAdmin || claims.Role == domain.RoleSuperAdmin
	if booking.UserID != claims.UserID && !isStaff {
		// 404 instead of 403, so PNRs can't be probed
		bookingWriteError(w, http.StatusNotFound, "booking not found")
		return
	}

	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	bookingWriteJSON(w, http.StatusOK, map[string]any{
		"booking": booking,
		"tickets": tickets,
	})
}

// MyBookings handles GET /api/v1/bookings/my
func (h *BookingHandler) MyBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookings, err := h.svc.GetUserBookings(r.Context(), claims.UserID)
	if err != nil {
		bookingHandleError(w, "MY BOOKINGS", err)
		return
	}

	if bookings == nil {
		bookings = []domain.Booking{}
	}
	bookingWriteJSON(w, http.StatusOK, bookings)
}
