package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/service"
)

type PaymentHandler struct {
	svc service.BookingService
}

func NewPaymentHandler(svc service.BookingService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// Initiate handles POST /api/v1/payments/initiate
func (h *PaymentHandler) Initiate(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req domain.InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	resp, err := h.svc.InitiatePayment(r.Context(), claims.UserID, &req)
	if err != nil {
		bookingHandleError(w, "INITIATE PAYMENT", err)
		return
	}

	bookingWriteJSON(w, http.StatusCreated, resp)
}

// Simulate handles POST /api/v1/payments/{paymentID}/simulate
// Body: {"result": "success" | "fail" | "cancel"}
func (h *PaymentHandler) Simulate(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req domain.SimulatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	resp, err := h.svc.SimulatePayment(r.Context(), claims.UserID, chi.URLParam(r, "paymentID"), req.Result)
	if err != nil {
		bookingHandleError(w, "SIMULATE PAYMENT", err)
		return
	}

	bookingWriteJSON(w, http.StatusOK, resp)
}

// Get handles GET /api/v1/payments/{paymentID}
func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.svc.GetPayment(r.Context(), claims.UserID, chi.URLParam(r, "paymentID"))
	if err != nil {
		bookingHandleError(w, "GET PAYMENT", err)
		return
	}

	bookingWriteJSON(w, http.StatusOK, resp)
}
