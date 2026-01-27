package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTokenNotFound = errors.New("reset token not found")
	ErrTokenExpired  = errors.New("reset token has expired")
	ErrTokenUsed     = errors.New("reset token has already been used")
)

// PasswordResetToken represents a password reset token entity
type PasswordResetToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// PasswordResetRepository defines interface for password reset token operations
type PasswordResetRepository interface {
	Create(ctx context.Context, userID string, expiresIn time.Duration) (*PasswordResetToken, error)
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID string) error
}

// passwordResetRepository implements PasswordResetRepository interface
type passwordResetRepository struct {
	db *sql.DB
}

// NewPasswordResetRepository creates new password reset repository instance
func NewPasswordResetRepository(db *sql.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create creates a new password reset token
func (r *passwordResetRepository) Create(ctx context.Context, userID string, expiresIn time.Duration) (*PasswordResetToken, error) {
	// Generate secure token (32 bytes = 64 hex characters)
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(expiresIn)

	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	resetToken := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	err = r.db.QueryRowContext(ctx, query, userID, token, expiresAt).
		Scan(&resetToken.ID, &resetToken.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create reset token: %w", err)
	}

	return resetToken, nil
}

// GetByToken retrieves a password reset token by its token value
func (r *passwordResetRepository) GetByToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	resetToken := &PasswordResetToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.Token,
		&resetToken.ExpiresAt,
		&resetToken.Used,
		&resetToken.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTokenNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get reset token: %w", err)
	}

	// Check if token is used
	if resetToken.Used {
		return nil, ErrTokenUsed
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return resetToken, nil
}

// MarkAsUsed marks a token as used
func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = TRUE
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// DeleteExpired deletes all expired tokens (cleanup job)
func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < NOW() OR used = TRUE
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all tokens for a user (invalidate previous tokens)
func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	return nil
}
