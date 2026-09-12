package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db         *sqlx.DB
	jwtService *JWTService
	supabase   *SupabaseAuthService
	cloud      *CloudAuthService
}

// refreshTokenDigest stores only a one-way digest in the database. A database
// read alone must not be enough to replay a refresh token.
func refreshTokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", digest[:])
}

func isMissingRefreshTokenTable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "refresh_tokens") &&
		(strings.Contains(message, "does not exist") || strings.Contains(message, "no such table"))
}

// persistRefreshToken returns false only when running against a legacy
// database that has not received the refresh-token migration yet. This keeps
// old installations compatible while enabling revocation as soon as the
// migration is applied.
func (s *Service) persistRefreshToken(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), userID, refreshTokenDigest(token), time.Now().Add(s.jwtService.refreshTokenT), time.Now())
	if isMissingRefreshTokenTable(err) {
		log.Printf("refresh token revocation is disabled until refresh_tokens migration is applied")
		return false, nil
	}
	return true, err
}

// checkPersistedRefreshToken returns whether the store exists and whether the
// supplied token is currently registered for the user.
func (s *Service) checkPersistedRefreshToken(ctx context.Context, userID uuid.UUID, token string) (available, exists bool, err error) {
	var marker int
	err = s.db.QueryRowContext(ctx, `
		SELECT 1 FROM refresh_tokens WHERE user_id = $1 AND token = $2 LIMIT 1
	`, userID, refreshTokenDigest(token)).Scan(&marker)
	if isMissingRefreshTokenTable(err) {
		return false, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true, false, nil
	}
	if err != nil {
		return true, false, err
	}
	return true, true, nil
}

func (s *Service) revokeRefreshToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, refreshTokenDigest(token))
	if isMissingRefreshTokenTable(err) {
		return nil
	}
	return err
}

// IsSubscriptionExpired reports whether the user's subscription is no longer valid.
func (s *Service) IsSubscriptionExpired(subscriptionStatus string, expiresAt *time.Time) bool {
	normalizedStatus := strings.ToLower(strings.TrimSpace(subscriptionStatus))
	switch normalizedStatus {
	case "canceled", "cancelled", "expired", "deleted":
		return true
	}

	if expiresAt != nil && time.Now().UTC().After(expiresAt.UTC()) {
		return true
	}

	return false
}

// checkSubscriptionStatus checks if user's subscription is valid (from worktrack)
func (s *Service) checkSubscriptionStatus(subscriptionStatus string, expiresAt *time.Time) error {
	normalizedStatus := strings.ToLower(strings.TrimSpace(subscriptionStatus))
	if s.IsSubscriptionExpired(normalizedStatus, expiresAt) {
		if normalizedStatus == "canceled" || normalizedStatus == "cancelled" || normalizedStatus == "deleted" {
			return errors.New("subscription canceled")
		}
		return errors.New("subscription expired")
	}

	return nil
}

// validatePassword checks password with bcrypt and PostgreSQL crypt fallback (from worktrack)
func (s *Service) validatePassword(password, storedHash, email string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return true
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_MODE")))
	if strings.HasPrefix(databaseURL, "sqlite://") || mode == "local" {
		return false
	}

	var passwordMatches bool
	err := s.db.QueryRow(`SELECT crypt($1, password_hash) = password_hash FROM users WHERE email = $2`, password, email).Scan(&passwordMatches)
	if err != nil {
		log.Printf("Password fallback check failed for %s: %v", email, err)
		return false
	}

	return passwordMatches
}

func NewService(db *sqlx.DB, jwtSecret string, useSupabase bool, supabaseURL, supabaseKey, cloudAPIURL string) (*Service, error) {
	jwtService := NewJWTService(jwtSecret, 15*time.Minute, 7*24*time.Hour)

	var supabase *SupabaseAuthService
	var err error
	if useSupabase {
		supabase, err = NewSupabaseAuthService(supabaseURL, supabaseKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Supabase: %w", err)
		}
	}

	var cloud *CloudAuthService
	if cloudAPIURL != "" {
		cloud = NewCloudAuthService(cloudAPIURL)
	}

	return &Service{
		db:         db,
		jwtService: jwtService,
		supabase:   supabase,
		cloud:      cloud,
	}, nil
}

// ValidateCloudAccess is the local API's trust boundary for protected
// requests. Local SQLite credentials alone must never authorize operations.
func (s *Service) ValidateCloudAccess(ctx context.Context, cloudToken string) error {
	if s.cloud == nil {
		return errors.New("cloud authentication is not configured")
	}
	validation, err := s.cloud.ValidateCloudToken(ctx, strings.TrimSpace(cloudToken))
	if err != nil {
		return err
	}
	if !validation.cloudAccountIsActive() || s.IsSubscriptionExpiredFromCloud(validation.cloudSubscriptionStatus(), validation.cloudSubscriptionExpiresAt()) {
		return errors.New("cloud account is inactive or subscription is expired")
	}
	return nil
}

func (s *Service) IsSubscriptionExpiredFromCloud(status, expiresAt string) bool {
	if strings.EqualFold(strings.TrimSpace(status), "canceled") || strings.EqualFold(strings.TrimSpace(status), "cancelled") || strings.EqualFold(strings.TrimSpace(status), "expired") {
		return true
	}
	if expiresAt == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, expiresAt)
	return err == nil && time.Now().UTC().After(parsed.UTC())
}

// Register registers a new admin user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	var existingUser User
	err := s.db.GetContext(ctx, &existingUser, "SELECT id FROM users WHERE email = $1", req.Email)
	if err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create owner user with default subscription
	user := &User{
		ID:                 uuid.New(),
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
	expiresAt := time.Now().UTC().AddDate(1, 0, 0)
	user.SubscriptionExpiresAt = &expiresAt

	// Insert user with fallback for schema differences
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, is_active,
		                  subscription_status, subscription_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Phone, user.IsActive,
		user.SubscriptionStatus, user.SubscriptionExpiresAt,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens with user_id only
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if _, err := s.persistRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
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
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE email = $1 AND is_active = TRUE
	`

	var row userRow
	err := s.db.QueryRowxContext(ctx, query, req.Email).Scan(
		&row.ID,
		&row.Email,
		&row.PasswordHash,
		&row.FirstName,
		&row.LastName,
		&row.Phone,
		&row.IsActive,
		&row.LastLoginAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.SubscriptionStatus,
		&row.SubscriptionExpiresAt,
	)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	// Verify password with fallback support (from worktrack)
	if !s.validatePassword(req.Password, user.PasswordHash, req.Email) {
		return nil, ErrInvalidCredentials
	}

	// Check subscription status (from worktrack)
	if err := s.checkSubscriptionStatus(user.SubscriptionStatus, user.SubscriptionExpiresAt); err != nil {
		return nil, ErrUnauthorized
	}

	// Update last login
	now := time.Now()
	_, err = s.db.ExecContext(ctx, "UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3", now, now, user.ID)
	if err != nil {
		// Log error but don't fail login
		log.Printf("failed to update last login: %v", err)
	}

	// Generate tokens with user_id only
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if _, err := s.persistRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
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
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	storeAvailable, tokenExists, err := s.checkPersistedRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("check refresh token: %w", err)
	}
	if storeAvailable && !tokenExists {
		return nil, ErrInvalidToken
	}

	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE id = $1
	`

	var row userRow
	err = s.db.QueryRowxContext(ctx, query, claims.UserID).Scan(
		&row.ID,
		&row.Email,
		&row.PasswordHash,
		&row.FirstName,
		&row.LastName,
		&row.Phone,
		&row.IsActive,
		&row.LastLoginAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.SubscriptionStatus,
		&row.SubscriptionExpiresAt,
	)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
	}

	if err := s.checkSubscriptionStatus(user.SubscriptionStatus, user.SubscriptionExpiresAt); err != nil {
		return nil, ErrUnauthorized
	}

	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	// Generate new access token
	newAccessToken, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	// Rotate the refresh token when durable storage is available. This makes a
	// stolen token single-use after a successful refresh and lets Logout revoke
	// all remaining sessions from the cloud database.
	nextRefreshToken := refreshToken
	if storeAvailable {
		nextRefreshToken, err = s.jwtService.GenerateRefreshToken(user.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
		}
		if _, err := s.persistRefreshToken(ctx, user.ID, nextRefreshToken); err != nil {
			return nil, fmt.Errorf("failed to persist rotated refresh token: %w", err)
		}
		if err := s.revokeRefreshToken(ctx, refreshToken); err != nil {
			return nil, fmt.Errorf("failed to revoke previous refresh token: %w", err)
		}
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: nextRefreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         user,
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *Service) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	return s.jwtService.ValidateToken(token)
}

// CreateCloudSession creates a local JWT session from a validated cloud token
func (s *Service) CreateCloudSession(ctx context.Context, cloudToken string) (*CloudSessionResponse, error) {
	if s.cloud == nil {
		return nil, fmt.Errorf("cloud auth is not configured")
	}

	return s.cloud.CreateLocalSession(ctx, s.jwtService, s.db, cloudToken)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var row userRow
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE id = $1
	`

	err := s.db.GetContext(ctx, &row, query, userID.String())
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, ErrUserNotFound
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
		return ErrInvalidPassword
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
	_, err := s.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if isMissingRefreshTokenTable(err) {
		return nil
	}
	return err
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

	query := `
		SELECT user_id 
		FROM password_reset_tokens 
		WHERE token = $1 AND used = FALSE AND expires_at > NOW()
	`

	err := s.db.GetContext(ctx, &userID, query, token)
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
