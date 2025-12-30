package entity

import "time"

// User represents user data from auth service
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	FullName  string    `db:"full_name"`
	Phone     string    `db:"phone"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// User role constants
const (
	UserRoleCustomer  = "customer"
	UserRoleOrganizer = "organizer"
	UserRoleAdmin     = "admin"
)

// IsCustomer checks if user is a customer
func (u *User) IsCustomer() bool {
	return u.Role == UserRoleCustomer
}

// IsOrganizer checks if user is an organizer
func (u *User) IsOrganizer() bool {
	return u.Role == UserRoleOrganizer
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
