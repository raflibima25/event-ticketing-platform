package entity

import "time"

// User represents the user entity in database
type User struct {
	ID              string    `json:"id" db:"id"`
	Email           string    `json:"email" db:"email"`
	PasswordHash    string    `json:"-" db:"password_hash"` // Never expose password in JSON
	FullName        string    `json:"full_name" db:"full_name"`
	Phone           *string   `json:"phone,omitempty" db:"phone"`
	Role            string    `json:"role" db:"role"` // customer, organizer, admin
	IsEmailVerified bool      `json:"is_email_verified" db:"is_email_verified"`
	OAuthProvider   *string   `json:"oauth_provider,omitempty" db:"oauth_provider"`
	OAuthID         *string   `json:"oauth_id,omitempty" db:"oauth_id"`
	IsDeleted       bool      `json:"-" db:"is_deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole constants
const (
	RoleCustomer  = "customer"
	RoleOrganizer = "organizer"
	RoleAdmin     = "admin"
)

// IsValidRole checks if role is valid
func IsValidRole(role string) bool {
	switch role {
	case RoleCustomer, RoleOrganizer, RoleAdmin:
		return true
	default:
		return false
	}
}
