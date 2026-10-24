package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/service"
)

// AdminTripHandler serves the demo admin screens: vehicles, routes, trips and trip creation.
type AdminTripHandler struct {
	svc *service.AdminService
}

func NewAdminTripHandler(svc *service.AdminService) *AdminTripHandler {
	return &AdminTripHandler{svc: svc}
}

func adminHandleError(w http.ResponseWriter, tag string, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		bookingWriteError(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, service.ErrAdminForbidden):
		bookingWriteError(w, http.StatusForbidden, "you do not have access to this resource")
	case errors.Is(err, service.ErrAdminConflict):
		bookingWriteError(w, http.StatusConflict, strings.TrimPrefix(err.Error(), "conflict: "))
	default:
		log.Printf("[%s ERROR] %+v", tag, err)
		bookingWriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ListVehicles handles GET /api/v1/admin/vehicles
func (h *AdminTripHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vehicles, err := h.svc.ListVehicles(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		adminHandleError(w, "ADMIN VEHICLES", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, vehicles)
}

// ListRoutes handles GET /api/v1/admin/routes
func (h *AdminTripHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.svc.ListRoutes(r.Context())
	if err != nil {
		adminHandleError(w, "ADMIN ROUTES", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, routes)
}

// ListTrips handles GET /api/v1/admin/trips (upcoming trips, limited to the caller's operator)
func (h *AdminTripHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	trips, err := h.svc.ListTrips(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		adminHandleError(w, "ADMIN TRIPS", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, trips)
}

// CreateTrip handles POST /api/v1/admin/trips
// Body: {"vehicle_id","route_id","departure_time","arrival_time","base_fare"}
func (h *AdminTripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req domain.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	resp, err := h.svc.CreateTrip(r.Context(), claims.UserID, claims.Role, &req)
	if err != nil {
		adminHandleError(w, "ADMIN CREATE TRIP", err)
		return
	}
	bookingWriteJSON(w, http.StatusCreated, resp)
}

// ListOperators handles GET /api/v1/admin/operators
func (h *AdminTripHandler) ListOperators(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	operators, err := h.svc.ListOperators(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		adminHandleError(w, "ADMIN OPERATORS", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, operators)
}

// CreateRoute handles POST /api/v1/admin/routes
// Body: {"origin_city","destination_city"}
func (h *AdminTripHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<14)

	var req domain.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	route, err := h.svc.CreateRoute(r.Context(), claims.UserID, claims.Role, &req)
	if err != nil {
		adminHandleError(w, "ADMIN CREATE ROUTE", err)
		return
	}
	bookingWriteJSON(w, http.StatusCreated, route)
}

// CreateVehicle handles POST /api/v1/admin/vehicles
// Body: {"operator_id"?,"registration_no","type","bus_layout"?,"coaches"?}
func (h *AdminTripHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req domain.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	vehicle, err := h.svc.CreateVehicle(r.Context(), claims.UserID, claims.Role, &req)
	if err != nil {
		adminHandleError(w, "ADMIN CREATE VEHICLE", err)
		return
	}
	bookingWriteJSON(w, http.StatusCreated, vehicle)
}

// CreateOperator handles POST /api/v1/admin/operators (super admin only)
func (h *AdminTripHandler) CreateOperator(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<14)

	var req domain.CreateOperatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	op, err := h.svc.CreateOperator(r.Context(), claims.Role, &req)
	if err != nil {
		adminHandleError(w, "ADMIN CREATE OPERATOR", err)
		return
	}
	bookingWriteJSON(w, http.StatusCreated, op)
}

// ListOperatorAdmins handles GET /api/v1/admin/operator-admins (super admin only)
func (h *AdminTripHandler) ListOperatorAdmins(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	admins, err := h.svc.ListOperatorAdmins(r.Context(), claims.Role)
	if err != nil {
		adminHandleError(w, "ADMIN LIST OPERATOR ADMINS", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, admins)
}

// CreateOperatorAdmin handles POST /api/v1/admin/operator-admins (super admin only)
func (h *AdminTripHandler) CreateOperatorAdmin(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<14)

	var req domain.CreateOperatorAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bookingWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	res, err := h.svc.CreateOperatorAdmin(r.Context(), claims.Role, &req)
	if err != nil {
		adminHandleError(w, "ADMIN CREATE OPERATOR ADMIN", err)
		return
	}
	bookingWriteJSON(w, http.StatusCreated, res)
}

// DeleteOperatorAdmin handles DELETE /api/v1/admin/operator-admins/{userID} (super admin only)
func (h *AdminTripHandler) DeleteOperatorAdmin(w http.ResponseWriter, r *http.Request) {
	claims, ok := bookingClaims(r)
	if !ok {
		bookingWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.RemoveOperatorAdmin(r.Context(), claims.Role, chi.URLParam(r, "userID")); err != nil {
		adminHandleError(w, "ADMIN REMOVE OPERATOR ADMIN", err)
		return
	}
	bookingWriteJSON(w, http.StatusOK, map[string]string{"message": "operator admin removed"})
}
