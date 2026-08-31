package auth

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system (owner only)
type User struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	Email                 string     `json:"email" db:"email"`
	PasswordHash          string     `json:"-" db:"password_hash"`
	FirstName             string     `json:"first_name" db:"first_name"`
	LastName              string     `json:"last_name" db:"last_name"`
	Phone                 string     `json:"phone" db:"phone"`
	IsActive              bool       `json:"is_active" db:"is_active"`
	LastLoginAt           *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	SubscriptionStatus    string     `json:"subscription_status" db:"subscription_status"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at" db:"subscription_expires_at"`
}

type userRow struct {
	ID                    string         `db:"id"`
	Email                 string         `db:"email"`
	PasswordHash          string         `db:"password_hash"`
	FirstName             string         `db:"first_name"`
	LastName              string         `db:"last_name"`
	Phone                 sql.NullString `db:"phone"`
	IsActive              bool           `db:"is_active"`
	LastLoginAt           sql.NullString `db:"last_login_at"`
	CreatedAt             string         `db:"created_at"`
	UpdatedAt             string         `db:"updated_at"`
	SubscriptionStatus    sql.NullString `db:"subscription_status"`
	SubscriptionExpiresAt sql.NullString `db:"subscription_expires_at"`
}

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}

	if idx := strings.Index(trimmed, " m="); idx > 0 {
		trimmed = strings.TrimSpace(trimmed[:idx])
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999 -0700",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported SQLite timestamp format: %q", raw)
}

func userFromRow(row userRow) (User, error) {
	userID, err := uuid.Parse(row.ID)
	if err != nil {
		return User{}, fmt.Errorf("parse user id %q: %w", row.ID, err)
	}

	user := User{
		ID:                 userID,
		Email:              row.Email,
		PasswordHash:       row.PasswordHash,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		IsActive:           row.IsActive,
		SubscriptionStatus: "active",
	}
	if row.Phone.Valid {
		user.Phone = row.Phone.String
	}
	if row.SubscriptionStatus.Valid && strings.TrimSpace(row.SubscriptionStatus.String) != "" {
		user.SubscriptionStatus = row.SubscriptionStatus.String
	}
	if row.CreatedAt != "" {
		parsedCreatedAt, err := parseSQLiteTimestamp(row.CreatedAt)
		if err != nil {
			return User{}, err
		}
		user.CreatedAt = parsedCreatedAt
	}
	if row.UpdatedAt != "" {
		parsedUpdatedAt, err := parseSQLiteTimestamp(row.UpdatedAt)
		if err != nil {
			return User{}, err
		}
		user.UpdatedAt = parsedUpdatedAt
	}
	if row.LastLoginAt.Valid && strings.TrimSpace(row.LastLoginAt.String) != "" {
		parsedLoginAt, err := parseSQLiteTimestamp(row.LastLoginAt.String)
		if err == nil {
			user.LastLoginAt = &parsedLoginAt
		}
	}
	if row.SubscriptionExpiresAt.Valid && strings.TrimSpace(row.SubscriptionExpiresAt.String) != "" {
		parsedExpiresAt, err := parseSQLiteTimestamp(row.SubscriptionExpiresAt.String)
		if err == nil {
			user.SubscriptionExpiresAt = &parsedExpiresAt
		}
	}
	return user, nil
}

// RefreshToken represents a refresh token for JWT
type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents password reset confirmation
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
