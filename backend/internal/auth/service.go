package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db         *sqlx.DB
	jwtService *JWTService
	supabase   *SupabaseAuthService
}

// checkSubscriptionStatus checks if user's subscription is valid (from worktrack)
func (s *Service) checkSubscriptionStatus(subscriptionStatus string, expiresAt sql.NullTime) error {
	if subscriptionStatus == "canceled" {
		return errors.New("subscription canceled")
	}

	if subscriptionStatus == "expired" {
		return errors.New("subscription expired")
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return errors.New("subscription expired")
	}

	return nil
}

// validatePassword checks password with bcrypt and PostgreSQL crypt fallback (from worktrack)
func (s *Service) validatePassword(password, storedHash, email string) bool {
	// Try bcrypt first
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return true
	}

	// Fallback to PostgreSQL crypt function
	var passwordMatches bool
	err := s.db.QueryRow(`SELECT crypt($1, password_hash) = password_hash FROM users WHERE email = $2`, password, email).Scan(&passwordMatches)
	if err != nil {
		log.Printf("Password fallback check failed for %s: %v", email, err)
		return false
	}

	return passwordMatches
}

func NewService(db *sqlx.DB, jwtSecret string, useSupabase bool, supabaseURL, supabaseKey string) (*Service, error) {
	jwtService := NewJWTService(jwtSecret, 15*time.Minute, 7*24*time.Hour)
	
	var supabase *SupabaseAuthService
	var err error
	if useSupabase {
		supabase, err = NewSupabaseAuthService(supabaseURL, supabaseKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Supabase: %w", err)
		}
	}

	return &Service{
		db:         db,
		jwtService: jwtService,
		supabase:   supabase,
	}, nil
}

// Register registers a new admin user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	var existingUser User
	err := s.db.GetContext(ctx, &existingUser, "SELECT id FROM users WHERE email = $1", req.Email)
	if err == nil {
		return nil, fmt.Errorf("user already exists: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user with default subscription
	user := &User{
		ID:                 uuid.New(),
		OrganizationID:     req.OrganizationID,
		Email:              req.Email,
		PasswordHash:       string(hashedPassword),
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Phone:              req.Phone,
		IsActive:           true,
		SubscriptionStatus: "active",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Set default subscription expiry (1 year from now)
	expiresAt := time.Now().AddDate(1, 0, 0)
	user.SubscriptionExpiresAt = &expiresAt

	// Insert user
	query := `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, phone, is_active,
		                  subscription_status, subscription_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		user.ID, user.OrganizationID, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone, user.IsActive,
		user.SubscriptionStatus, user.SubscriptionExpiresAt,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get role name for token
	var roleName string
	err = s.db.GetContext(ctx, &roleName, "SELECT name FROM roles WHERE id = $1", user.RoleID)
	if err != nil {
		roleName = "admin" // default
	}

	// Generate tokens with user_id and role
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String(), roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         *user,
	}, nil
}

// Login authenticates an admin user (from worktrack)
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Get user by email with subscription info
	var user User
	var subscriptionExpiresAt sql.NullTime
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name,
		       phone, role_id, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE email = $1 AND is_active = TRUE
	`

	err := s.db.GetContext(ctx, &user, query, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password with fallback support (from worktrack)
	if !s.validatePassword(req.Password, user.PasswordHash, req.Email) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check subscription status (from worktrack)
	if err := s.checkSubscriptionStatus(user.SubscriptionStatus, sql.NullTime{
		Time:  func() time.Time {
			if user.SubscriptionExpiresAt != nil {
				return *user.SubscriptionExpiresAt
			}
			return time.Time{}
		}(),
		Valid: user.SubscriptionExpiresAt != nil,
	}); err != nil {
		return nil, fmt.Errorf("subscription error: %w", err)
	}

	// Update last login
	now := time.Now()
	_, err = s.db.ExecContext(ctx, "UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3", now, now, user.ID)
	if err != nil {
		// Log error but don't fail login
		log.Printf("failed to update last login: %v", err)
	}

	// Get role name for token
	var roleName string
	err = s.db.GetContext(ctx, &roleName, "SELECT name FROM roles WHERE id = $1", user.RoleID)
	if err != nil {
		roleName = "admin" // default
	}

	// Generate tokens with user_id and role (from worktrack)
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String(), roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         user,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	var user User
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name,
		       phone, role_id, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE id = $1
	`

	err = s.db.GetContext(ctx, &user, query, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new access token
	newAccessToken, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         user,
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *Service) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	return s.jwtService.ValidateToken(token)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name,
		       phone, role_id, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE id = $1
	`

	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	// Get current user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = s.db.ExecContext(ctx, 
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Logout handles user logout
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	// This would invalidate refresh tokens
	// For now, this is a placeholder
	return nil
}

// RequestPasswordReset initiates a password reset request
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// Check if user exists
	var user User
	err := s.db.GetContext(ctx, &user, "SELECT id, email FROM users WHERE email = $1 AND is_active = TRUE", email)
	if err != nil {
		// Don't reveal if user exists for security
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour) // Token valid for 1 hour

	// Store reset token
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET token = $3, expires_at = $4, created_at = NOW()
	`

	_, err = s.db.ExecContext(ctx, query, uuid.New(), user.ID, resetToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// In production, send email with reset link
	// For now, we'll just log the token
	log.Printf("Password reset token for %s: %s (valid until %s)", email, resetToken, expiresAt.Format(time.RFC3339))

	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate reset token
	var userID uuid.UUID
	var expiresAt time.Time
	
	query := `
		SELECT user_id, expires_at 
		FROM password_reset_tokens 
		WHERE token = $1 AND used = FALSE AND expires_at > NOW()
	`
	
	err := s.db.GetContext(ctx, &userID, &expiresAt, query, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	_, err = s.db.ExecContext(ctx, 
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	_, err = s.db.ExecContext(ctx, 
		"UPDATE password_reset_tokens SET used = TRUE, used_at = NOW() WHERE token = $1",
		token)
	if err != nil {
		log.Printf("failed to mark reset token as used: %v", err)
	}

	return nil
}
