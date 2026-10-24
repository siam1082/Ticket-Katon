package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/pkg/mailer"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/repository"
)

const (
	minPasswordLen    = 8
	maxPasswordBytes  = 72 // bcrypt limit
	verificationTTL   = 10 * time.Minute
	maxVerifyAttempts = 5
	resendCooldown    = 60 * time.Second
)

var (
	// Bangladeshi mobile numbers: 01 + operator digit (3-9) + 8 digits.
	bdPhoneRe = regexp.MustCompile(`^01[3-9][0-9]{8}$`)
	sixDigits = regexp.MustCompile(`^[0-9]{6}$`)

	ErrEmailNotVerified = errors.New("email not verified")
	ErrInvalidCode      = errors.New("invalid or expired code")
	ErrTooManyAttempts  = errors.New("too many wrong codes")
)

// dummyHash is compared against when the email is unknown, so "no such user"
// and "wrong password" take about the same time.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// ValidationError is returned for bad client input. Handlers map it to HTTP 400.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

type AuthService struct {
	userRepo      repository.UserRepository
	tokenMaker    *token.TokenMaker
	mailer        mailer.Mailer
	exposeDevCode bool // development only: show the code in the API response when no real mail is sent
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenMaker *token.TokenMaker,
	m mailer.Mailer,
	exposeDevCode bool,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		tokenMaker:    tokenMaker,
		mailer:        m,
		exposeDevCode: exposeDevCode,
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

type VerifyEmailInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// RegisterResponse: the account exists but cannot sign in until the email is verified.
type RegisterResponse struct {
	Message              string `json:"message"`
	Email                string `json:"email"`
	RequiresVerification bool   `json:"requires_verification"`
	EmailSent            bool   `json:"email_sent"`
	DevCode              string `json:"dev_code,omitempty"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// normalizeBDPhone accepts 01XXXXXXXXX, +8801XXXXXXXXX and 8801XXXXXXXXX (spaces and dashes allowed)
// and returns the 01XXXXXXXXX form.
func normalizeBDPhone(s string) (string, bool) {
	s = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "+88"):
		s = s[3:]
	case strings.HasPrefix(s, "88") && len(s) == 13:
		s = s[2:]
	}
	if !bdPhoneRe.MatchString(s) {
		return "", false
	}
	return s, true
}

// passwordProblem returns why a password is not acceptable, or "" if it is fine.
func passwordProblem(p string) string {
	if len(p) < minPasswordLen {
		return "password must be at least 8 characters"
	}
	if len(p) > maxPasswordBytes {
		return "password must be at most 72 bytes"
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "password must contain at least one letter and one number"
	}
	return ""
}

// validateRegister trims and normalizes the input in place.
func validateRegister(in *RegisterInput) error {
	in.FullName = cleanSpaces(in.FullName)
	in.Email = normalizeEmail(in.Email)

	if n := utf8.RuneCountInString(in.FullName); n < 2 || n > 100 {
		return &ValidationError{"full name must be 2 to 100 characters"}
	}

	addr, err := mail.ParseAddress(in.Email)
	if err != nil || addr.Address != in.Email || len(in.Email) > 254 {
		return &ValidationError{"please enter a valid email address"}
	}
	if isDisposableEmail(in.Email) {
		return &ValidationError{"temporary email addresses are not accepted, please use your own email"}
	}

	phone, ok := normalizeBDPhone(in.Phone)
	if !ok {
		return &ValidationError{"please enter a valid Bangladeshi mobile number, like 01712345678"}
	}
	in.Phone = phone

	if msg := passwordProblem(in.Password); msg != "" {
		return &ValidationError{msg}
	}
	return nil
}

func newVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashVerificationCode(userID, code string) string {
	sum := sha256.Sum256([]byte(userID + ":" + code))
	return hex.EncodeToString(sum[:])
}

// sendVerification stores a fresh code and emails it. It never fails the caller:
// if the email cannot be sent the user can ask for a new code.
func (s *AuthService) sendVerification(ctx context.Context, user *domain.User) (sent bool, devCode string) {
	code, err := newVerificationCode()
	if err != nil {
		log.Printf("[AUTH] could not create verification code: %v", err)
		return false, ""
	}
	if err := s.userRepo.CreateEmailVerification(ctx, user.ID, hashVerificationCode(user.ID, code), time.Now().Add(verificationTTL)); err != nil {
		log.Printf("[AUTH] could not store verification code: %v", err)
		return false, ""
	}

	body := fmt.Sprintf(
		"Hello %s,\n\nYour TicketKaton verification code is: %s\n\nIt expires in %d minutes. "+
			"If you did not create this account, you can ignore this email.\n\n"+
			"TicketKaton is an academic demo project.\n",
		user.FullName, code, int(verificationTTL.Minutes()))

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.mailer.Send(sendCtx, user.Email, "Your TicketKaton verification code", body); err != nil {
		log.Printf("[AUTH] could not send verification email to %s: %v", user.Email, err)
	} else {
		sent = true
	}

	// In development, with no real mail provider (or a failed send), show the code on screen.
	if s.exposeDevCode && (!s.mailer.Real() || !sent) {
		devCode = code
	}
	return sent, devCode
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*RegisterResponse, error) {
	if err := validateRegister(&in); err != nil {
		return nil, err
	}
	if !emailDomainCanReceiveMail(ctx, in.Email) {
		return nil, &ValidationError{"this email domain cannot receive mail, please check the address"}
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

	sent, devCode := s.sendVerification(ctx, user)

	msg := "Account created. We sent a 6-digit verification code to your email."
	if !sent {
		msg = "Account created, but we could not send the email. Use \"Send a new code\" to try again."
	}
	return &RegisterResponse{
		Message:              msg,
		Email:                user.Email,
		RequiresVerification: true,
		EmailSent:            sent,
		DevCode:              devCode,
	}, nil
}

// VerifyEmail checks the code, marks the account verified and signs the user in.
func (s *AuthService) VerifyEmail(ctx context.Context, in VerifyEmailInput) (*AuthResponse, error) {
	email := normalizeEmail(in.Email)
	code := strings.TrimSpace(in.Code)
	if email == "" || !sixDigits.MatchString(code) {
		return nil, ErrInvalidCode
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidCode
		}
		return nil, err
	}
	if user.EmailVerified {
		return nil, ErrInvalidCode // already verified: no code can be used to sign in
	}

	v, err := s.userRepo.GetLatestEmailVerification(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidCode
		}
		return nil, err
	}
	if time.Now().After(v.ExpiresAt) {
		return nil, ErrInvalidCode
	}
	if v.Attempts >= maxVerifyAttempts {
		return nil, ErrTooManyAttempts
	}

	want := hashVerificationCode(user.ID, code)
	if subtle.ConstantTimeCompare([]byte(want), []byte(v.CodeHash)) != 1 {
		_ = s.userRepo.IncrementVerificationAttempts(ctx, v.ID)
		return nil, ErrInvalidCode
	}

	if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, err
	}
	_ = s.userRepo.DeleteEmailVerifications(ctx, user.ID)
	user.EmailVerified = true

	jwtToken, err := s.tokenMaker.CreateToken(user, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: jwtToken, User: user}, nil
}

// ResendCode sends a new code if the account is waiting for verification. It always looks
// the same from outside, so it cannot be used to find out which emails are registered.
func (s *AuthService) ResendCode(ctx context.Context, email string) (devCode string, err error) {
	user, err := s.userRepo.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	if user.EmailVerified {
		return "", nil
	}

	if v, err := s.userRepo.GetLatestEmailVerification(ctx, user.ID); err == nil && time.Since(v.CreatedAt) < resendCooldown {
		return "", nil // asked again too soon
	}

	_, devCode = s.sendVerification(ctx, user)
	return devCode, nil
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

	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
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
