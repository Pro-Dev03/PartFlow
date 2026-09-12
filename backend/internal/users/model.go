package users

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	Email                 string     `json:"email" db:"email"`
	PasswordHash          string     `json:"-" db:"password_hash"`
	FirstName             string     `json:"first_name" db:"first_name"`
	LastName              string     `json:"last_name" db:"last_name"`
	Phone                 *string    `json:"phone,omitempty" db:"phone"`
	AvatarURL             *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive              bool       `json:"is_active" db:"is_active"`
	IsVerified            bool       `json:"is_verified" db:"is_verified"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	SubscriptionStatus    string     `json:"subscription_status" db:"subscription_status"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty" db:"subscription_expires_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// NewUser creates a new User instance
func NewUser(email, passwordHash, firstName, lastName string) *User {
	now := time.Now().UTC()
	expiresAt := now.AddDate(1, 0, 0)
	return &User{
		ID:                    uuid.New(),
		Email:                 email,
		PasswordHash:          passwordHash,
		FirstName:             firstName,
		LastName:              lastName,
		IsActive:              true,
		IsVerified:            false,
		SubscriptionStatus:    "active",
		SubscriptionExpiresAt: &expiresAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                    u.ID,
		Email:                 u.Email,
		FirstName:             u.FirstName,
		LastName:              u.LastName,
		FullName:              u.FullName(),
		Phone:                 u.Phone,
		AvatarURL:             u.AvatarURL,
		IsActive:              u.IsActive,
		IsVerified:            u.IsVerified,
		LastLoginAt:           u.LastLoginAt,
		SubscriptionStatus:    u.SubscriptionStatus,
		SubscriptionExpiresAt: u.SubscriptionExpiresAt,
		CreatedAt:             u.CreatedAt,
		UpdatedAt:             u.UpdatedAt,
	}
}
