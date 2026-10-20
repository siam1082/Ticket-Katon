package service

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/repository"
)

const (
	minPasswordLen   = 6
	maxPasswordBytes = 72 // bcrypt limit
)

var phoneRe = regexp.MustCompile(`^\+?[0-9]{10,15}$`)

// dummyHash is compared against when the email is unknown, so "no such user"
// and "wrong password" take about the same time.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// ValidationError is returned for bad client input. Handlers map it to HTTP 400.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

type AuthService struct {
	userRepo   repository.UserRepository
	tokenMaker *token.TokenMaker
}

func NewAuthService(userRepo repository.UserRepository, tokenMaker *token.TokenMaker) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}
}

// RegisterInput intentionally has no Role field: public registration
// always creates a CUSTOMER, whatever the client sends.
type RegisterInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// validateRegister trims and normalizes the input in place.
func validateRegister(in *RegisterInput) error {
	in.FullName = strings.TrimSpace(in.FullName)
	in.Email = normalizeEmail(in.Email)
	in.Phone = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(in.Phone))

	if n := utf8.RuneCountInString(in.FullName); n < 2 || n > 100 {
		return &ValidationError{"full name must be 2 to 100 characters"}
	}

	addr, err := mail.ParseAddress(in.Email)
	if err != nil || addr.Address != in.Email || len(in.Email) > 254 {
		return &ValidationError{"please enter a valid email address"}
	}

	if !phoneRe.MatchString(in.Phone) {
		return &ValidationError{"please enter a valid mobile number (10 to 15 digits)"}
	}

	if len(in.Password) < minPasswordLen {
		return &ValidationError{"password must be at least 6 characters"}
	}
	if len(in.Password) > maxPasswordBytes {
		return &ValidationError{"password must be at most 72 bytes"}
	}

	return nil
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResponse, error) {
	if err := validateRegister(&in); err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName:     in.FullName,
		Email:        in.Email,
		Phone:        in.Phone,
		PasswordHash: string(hashed),
		Role:         domain.RoleCustomer, // never taken from the request
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	jwtToken, err := s.tokenMaker.CreateToken(user, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: jwtToken,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, normalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(in.Password))
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	jwtToken, err := s.tokenMaker.CreateToken(user, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: jwtToken,
		User:  user,
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}