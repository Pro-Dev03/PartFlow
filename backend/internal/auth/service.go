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
	dbutil "github.com/partflow/smart-store/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db         *sqlx.DB
	jwtService *JWTService
	supabase   *SupabaseAuthService
	cloud      *CloudAuthService
}

var ErrRegistrationDisabled = errors.New("public registration is disabled")
var ErrPasswordResetUnavailable = errors.New("password reset delivery is not configured")

// refreshTokenDigest stores only a one-way digest in the database. A database
// read alone must not be enough to replay a refresh token.
func refreshTokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", digest[:])
}

func isMissingRefreshTokenTable(err error) bool {
	return isMissingTable(err, "refresh_tokens")
}

func isMissingTable(err error, tableName string) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, strings.ToLower(tableName)) &&
		(strings.Contains(message, "does not exist") || strings.Contains(message, "no such table"))
}

func lockUserForSession(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, userID uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at, session_version
		FROM users WHERE id = $1
	`
	if !strings.EqualFold(db.DriverName(), "sqlite") && !strings.EqualFold(db.DriverName(), "sqlite3") {
		query += ` FOR UPDATE`
	}
	var row userRow
	if err := tx.QueryRowxContext(ctx, query, userID).Scan(
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
		&row.SessionVersion,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
	}
	return &user, nil
}

// IsSubscriptionExpired reports whether the user's subscription is no longer valid.
func (s *Service) IsSubscriptionExpired(subscriptionStatus string, expiresAt *time.Time) bool {
	normalizedStatus := strings.ToLower(strings.TrimSpace(subscriptionStatus))
	switch normalizedStatus {
	case "suspended", "canceled", "cancelled", "expired", "deleted":
		return true
	case "active", "trial":
		// These are the only statuses that can authorize a session.
	default:
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
	if normalizedStatus == "suspended" {
		return ErrSubscriptionSuspended
	}
	if normalizedStatus == "deleted" {
		return ErrAccountDeleted
	}
	if s.IsSubscriptionExpired(normalizedStatus, expiresAt) {
		if normalizedStatus == "canceled" || normalizedStatus == "cancelled" {
			return errors.New("subscription canceled")
		}
		return ErrSubscriptionExpired
	}

	return nil
}

// validatePassword checks password with bcrypt and PostgreSQL crypt fallback (from worktrack)
func (s *Service) validatePassword(password, storedHash, email string) bool {
	return s.validatePasswordWithQuery(password, storedHash, email, s.db)
}

type passwordQueryer interface {
	QueryRow(query string, args ...any) *sql.Row
}

func (s *Service) validatePasswordWithQuery(password, storedHash, email string, queryer passwordQueryer) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return true
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_MODE")))
	if strings.HasPrefix(databaseURL, "sqlite://") || mode == "local" {
		return false
	}

	var passwordMatches bool
	err := queryer.QueryRow(`SELECT crypt($1, password_hash) = password_hash FROM users WHERE email = $2`, password, email).Scan(&passwordMatches)
	if err != nil {
		log.Printf("Password fallback check failed for %s: %v", email, err)
		return false
	}

	return passwordMatches
}

func NewService(db *sqlx.DB, jwtSecret string, useSupabase bool, supabaseURL, supabaseKey, cloudAPIURL string) (*Service, error) {
	jwtService := NewJWTService(jwtSecret, 15*time.Minute, 7*24*time.Hour)

	var supabase *SupabaseAuthService
	if useSupabase {
		return nil, errors.New("Supabase authentication is not implemented; set USE_SUPABASE_AUTH=false")
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
	// The cloud validate response is authoritative. Do not compare the cloud
	// expiry timestamp with the local machine clock, because a skewed desktop
	// clock must not end an otherwise valid subscription.
	if !validation.cloudAccountIsActive() {
		return errors.New("cloud account is inactive or subscription is expired")
	}
	return nil
}

func (s *Service) IsSubscriptionExpiredFromCloud(status, expiresAt string) bool {
	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	if normalizedStatus != "active" && normalizedStatus != "trial" {
		return true
	}
	if expiresAt == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, expiresAt)
	return err == nil && time.Now().UTC().After(parsed.UTC())
}

// Register is retained for compatibility with older callers, but user
// accounts are provisioned only through the authenticated administrator API.
// Public registration must never grant an active subscription.
func (s *Service) Register(_ context.Context, _ *RegisterRequest) (*AuthResponse, error) {
	return nil, ErrRegistrationDisabled
}

// Login authenticates an admin user (from worktrack)
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at
		FROM users WHERE email = $1
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("load user for login: %w", err)
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
	}

	// Check the password once before opening a transaction so bcrypt work does
	// not hold a database lock. The password and account state are checked again
	// under a row lock before any session token is committed.
	if !s.validatePassword(req.Password, user.PasswordHash, req.Email) {
		return nil, ErrInvalidCredentials
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin login session: %w", err)
	}
	defer tx.Rollback()

	lockedUser, err := lockUserForSession(ctx, tx, s.db, user.ID)
	if err != nil {
		return nil, fmt.Errorf("lock user for login: %w", err)
	}
	user = *lockedUser
	if !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(req.Email)) {
		return nil, ErrInvalidCredentials
	}
	if !s.validatePasswordWithQuery(req.Password, user.PasswordHash, user.Email, tx) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, user.ID); err != nil {
			if isMissingRefreshTokenTable(err) {
				return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
			}
			return nil, fmt.Errorf("revoke refresh tokens for inactive account: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit refresh-token revocation: %w", err)
		}
		return nil, ErrInactiveUser
	}
	if err := s.checkSubscriptionStatus(user.SubscriptionStatus, user.SubscriptionExpiresAt); err != nil {
		if _, revokeErr := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, user.ID); revokeErr != nil {
			if isMissingRefreshTokenTable(revokeErr) {
				return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", revokeErr)
			}
			return nil, fmt.Errorf("revoke refresh tokens for blocked account: %w", revokeErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit refresh-token revocation: %w", commitErr)
		}
		return nil, err
	}

	accessToken, err := s.jwtService.GenerateAccessTokenForSession(user.ID.String(), user.SessionVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshTokenForSession(user.ID.String(), user.SessionVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), user.ID, refreshTokenDigest(refreshToken), time.Now().Add(s.jwtService.refreshTokenT), time.Now()); err != nil {
		if isMissingRefreshTokenTable(err) {
			return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
		}
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE users SET last_login_at = $1 WHERE id = $2`, now, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit login session: %w", err)
	}
	user.LastLoginAt = &now

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         user,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate the signed token first, then lock the account and rotate its
	// persisted digest in one transaction. Password changes and logout update
	// the same account row before deleting refresh tokens, so exactly one side
	// of a concurrent refresh/revocation can commit.
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin refresh-token rotation: %w", err)
	}
	defer tx.Rollback()

	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       phone, is_active, last_login_at, created_at, updated_at,
		       subscription_status, subscription_expires_at, session_version
		FROM users WHERE id = $1
	`
	if !strings.EqualFold(s.db.DriverName(), "sqlite") && !strings.EqualFold(s.db.DriverName(), "sqlite3") {
		query += ` FOR UPDATE`
	}

	var row userRow
	err = tx.QueryRowxContext(ctx, query, userID).Scan(
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
		&row.SessionVersion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("load user for refresh: %w", err)
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
	}
	if claims.SessionVersion != user.SessionVersion {
		return nil, ErrInvalidToken
	}

	if err := s.checkSubscriptionStatus(user.SubscriptionStatus, user.SubscriptionExpiresAt); err != nil {
		if _, revokeErr := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, user.ID); revokeErr != nil {
			if isMissingRefreshTokenTable(revokeErr) {
				return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", revokeErr)
			}
			return nil, fmt.Errorf("revoke refresh tokens for blocked account: %w", revokeErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit refresh-token revocation: %w", commitErr)
		}
		return nil, err
	}

	if !user.IsActive {
		if _, revokeErr := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, user.ID); revokeErr != nil {
			if isMissingRefreshTokenTable(revokeErr) {
				return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", revokeErr)
			}
			return nil, fmt.Errorf("revoke refresh tokens for inactive account: %w", revokeErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit refresh-token revocation: %w", commitErr)
		}
		return nil, ErrInactiveUser
	}

	var marker int
	err = tx.QueryRowxContext(ctx, `
		SELECT 1 FROM refresh_tokens WHERE user_id = $1 AND token = $2 LIMIT 1
	`, user.ID, refreshTokenDigest(refreshToken)).Scan(&marker)
	if isMissingRefreshTokenTable(err) {
		return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("check refresh token: %w", err)
	}

	newAccessToken, err := s.jwtService.GenerateAccessTokenForSession(user.ID.String(), user.SessionVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	nextRefreshToken, err := s.jwtService.GenerateRefreshTokenForSession(user.ID.String(), user.SessionVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET token = $3, expires_at = $4, created_at = $5
		WHERE user_id = $1 AND token = $2
	`, user.ID, refreshTokenDigest(refreshToken), refreshTokenDigest(nextRefreshToken),
		time.Now().Add(s.jwtService.refreshTokenT), time.Now())
	if err != nil {
		if isMissingRefreshTokenTable(err) {
			return nil, fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
		}
		return nil, fmt.Errorf("rotate persisted refresh token: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return nil, ErrInvalidToken
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit refresh-token rotation: %w", err)
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
		       subscription_status, subscription_expires_at, session_version
		FROM users WHERE id = $1
	`

	err := s.db.GetContext(ctx, &row, query, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("load user by ID: %w", err)
	}

	user, err := userFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("read user profile: %w", err)
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
	if !s.validatePassword(req.CurrentPassword, user.PasswordHash, user.Email) {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password and revoke every session atomically.
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx,
		fmt.Sprintf("UPDATE users SET password_hash = $1, session_version = session_version + 1, updated_at = %s WHERE id = $2", dbutil.NowSQL(s.db)),
		string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrUserNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("revoke refresh tokens after password change: %w", err)
	}
	return tx.Commit()
}

// Logout handles user logout
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin logout revocation: %w", err)
	}
	defer tx.Rollback()
	if _, err := lockUserForSession(ctx, tx, s.db, userID); err != nil {
		return fmt.Errorf("lock user for logout: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET session_version = session_version + 1 WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("revoke access sessions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID); err != nil {
		if isMissingRefreshTokenTable(err) {
			return fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
		}
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}
	return tx.Commit()
}

// RequestPasswordReset initiates a password reset request
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// There is no verified email delivery path. Never create a reset credential
	// that would need to be exposed through application logs or an unsafe API.
	return ErrPasswordResetUnavailable
}

// ResetPassword resets a user's password using a reset token
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer tx.Rollback()

	resetQuery := `SELECT user_id FROM password_reset_tokens WHERE token = $1 AND used = FALSE AND expires_at > ` + dbutil.NowSQL(s.db)
	if !strings.EqualFold(s.db.DriverName(), "sqlite") {
		resetQuery += ` FOR UPDATE`
	}
	var userIDRaw string
	if err := tx.QueryRowxContext(ctx, resetQuery, token).Scan(&userIDRaw); err != nil {
		if isMissingTable(err, "password_reset_tokens") {
			return fmt.Errorf("password reset storage is unavailable: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("check password reset token: %w", err)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return fmt.Errorf("invalid password reset account")
	}

	lockUserQuery := `SELECT id FROM users WHERE id = $1`
	if !strings.EqualFold(s.db.DriverName(), "sqlite") {
		lockUserQuery += ` FOR UPDATE`
	}
	var lockedUserID string
	if err := tx.QueryRowxContext(ctx, lockUserQuery, userID).Scan(&lockedUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("lock password reset account: %w", err)
	}
	updateUserQuery := fmt.Sprintf(`UPDATE users SET password_hash = $1, session_version = session_version + 1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db))
	if _, err := tx.ExecContext(ctx, updateUserQuery, string(hashedPassword), userID); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID); err != nil {
		if isMissingRefreshTokenTable(err) {
			return fmt.Errorf("refresh token storage is unavailable: apply the refresh_tokens migration: %w", err)
		}
		return fmt.Errorf("revoke refresh tokens after password reset: %w", err)
	}
	markResetUsedQuery := fmt.Sprintf(`
		UPDATE password_reset_tokens
		SET used = TRUE, used_at = %s
		WHERE token = $1 AND used = FALSE AND expires_at > %s
	`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
	result, err := tx.ExecContext(ctx, markResetUsedQuery, token)
	if err != nil {
		return fmt.Errorf("failed to consume password reset token: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return fmt.Errorf("invalid or expired reset token")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit password reset: %w", err)
	}

	return nil
}
