package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/repository"
)

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

type RegisterInput struct {
	FullName string          `json:"full_name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Password string          `json:"password"`
	Role     domain.UserRole `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResponse, error) {
	if in.Role == "" {
		in.Role = domain.RoleCustomer
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
		Role:         in.Role,
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
	user, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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