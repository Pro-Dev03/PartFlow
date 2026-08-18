package users

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	FirstName     string     `json:"first_name" db:"first_name"`
	LastName      string     `json:"last_name" db:"last_name"`
	Phone         *string    `json:"phone,omitempty" db:"phone"`
	AvatarURL     *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	RoleID        *uuid.UUID `json:"role_id,omitempty" db:"role_id"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	IsVerified    bool       `json:"is_verified" db:"is_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// NewUser creates a new User instance
func NewUser(organizationID uuid.UUID, email, passwordHash, firstName, lastName string) *User {
	return &User{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Email:          email,
		PasswordHash:   passwordHash,
		FirstName:      firstName,
		LastName:       lastName,
		IsActive:       true,
		IsVerified:     false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		OrganizationID: u.OrganizationID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		FullName:      u.FullName(),
		Phone:         u.Phone,
		AvatarURL:     u.AvatarURL,
		RoleID:        u.RoleID,
		IsActive:      u.IsActive,
		IsVerified:    u.IsVerified,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}