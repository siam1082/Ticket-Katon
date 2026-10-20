package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/middleware"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/service"
)

const (
	maxLoginFailures = 5
	loginBlockWindow = 15 * time.Minute
)

type AuthHandler struct {
	svc     *service.AuthService
	limiter *loginLimiter
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		svc:     svc,
		limiter: newLoginLimiter(maxLoginFailures, loginBlockWindow),
	}
}

func authWriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func authWriteError(w http.ResponseWriter, status int, msg string) {
	authWriteJSON(w, status, map[string]string{"error": msg})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		authWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	resp, err := h.svc.Register(r.Context(), input)
	if err != nil {
		var ve *service.ValidationError
		switch {
		case errors.As(err, &ve):
			authWriteError(w, http.StatusBadRequest, ve.Msg)
		case errors.Is(err, domain.ErrConflict):
			authWriteError(w, http.StatusConflict, "email or mobile number is already registered")
		default:
			log.Printf("[REGISTER ERROR] %+v", err)
			authWriteError(w, http.StatusInternalServerError, "failed to create account")
		}
		return
	}

	authWriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		authWriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(input.Email) == "" || input.Password == "" {
		authWriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	key := strings.ToLower(strings.TrimSpace(input.Email)) + "|" + clientIP(r)
	if remaining, blocked := h.limiter.blocked(key); blocked {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(remaining.Seconds())+1))
		authWriteError(w, http.StatusTooManyRequests, "too many failed attempts, please try again later")
		return
	}

	resp, err := h.svc.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			h.limiter.fail(key)
			authWriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Printf("[LOGIN ERROR] %+v", err)
		authWriteError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	h.limiter.reset(key)
	authWriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		authWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		authWriteError(w, http.StatusNotFound, "user not found")
		return
	}

	authWriteJSON(w, http.StatusOK, user)
}
