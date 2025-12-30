package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// User represents user data from database
type User struct {
	ID        string
	Email     string
	FullName  string
	Phone     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository defines interface for user data operations
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates new user repository instance
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID retrieves user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, full_name, phone, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
