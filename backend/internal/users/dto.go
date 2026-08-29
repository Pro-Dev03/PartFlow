package users

import (
	"time"

	"github.com/google/uuid"
)

// UserRequest represents user creation/update request
type UserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password,omitempty"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	IsActive  bool    `json:"is_active"`
}

// UserResponse represents user response
type UserResponse struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	FullName              string     `json:"full_name"`
	Phone                 *string    `json:"phone,omitempty"`
	AvatarURL             *string    `json:"avatar_url,omitempty"`
	IsActive              bool       `json:"is_active"`
	IsVerified            bool       `json:"is_verified"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
	SubscriptionStatus    string     `json:"subscription_status"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// UserListRequest represents user list query parameters
type UserListRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PerPage   int    `form:"per_page" binding:"min=1,max=100"`
	Search    string `form:"search"`
	IsActive  *bool  `form:"is_active"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
