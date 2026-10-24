package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (full_name, email, phone_number, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		u.FullName,
		u.Email,
		u.Phone,
		u.PasswordHash,
		u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, full_name, email, phone_number, password_hash, role, created_at, updated_at, email_verified_at
		FROM users
		WHERE lower(email) = lower($1)
	`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EmailVerifiedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.EmailVerified = u.EmailVerifiedAt != nil
	return u, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, full_name, email, phone_number, password_hash, role, created_at, updated_at, email_verified_at
		FROM users
		WHERE id = $1
	`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EmailVerifiedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.EmailVerified = u.EmailVerifiedAt != nil
	return u, nil
}

func (r *PostgresUserRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email verified: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) CreateEmailVerification(ctx context.Context, userID, codeHash string, expiresAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to clear old codes: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO email_verifications (user_id, code_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, codeHash, expiresAt); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *PostgresUserRepository) GetLatestEmailVerification(ctx context.Context, userID string) (*repository.EmailVerification, error) {
	var v repository.EmailVerification
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, code_hash, expires_at, created_at, attempts
		FROM email_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, userID).Scan(&v.ID, &v.UserID, &v.CodeHash, &v.ExpiresAt, &v.CreatedAt, &v.Attempts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *PostgresUserRepository) IncrementVerificationAttempts(ctx context.Context, verificationID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE email_verifications SET attempts = attempts + 1 WHERE id = $1`, verificationID)
	return err
}

func (r *PostgresUserRepository) DeleteEmailVerifications(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID)
	return err
}
