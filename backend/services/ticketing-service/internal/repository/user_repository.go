package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository defines interface for user data operations
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates new user repository instance
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID retrieves user by ID using sqlx
func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	query := `
		SELECT id, email, full_name, phone, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
