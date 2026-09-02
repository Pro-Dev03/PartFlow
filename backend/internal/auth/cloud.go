package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CloudAuthService handles cloud token validation and local session creation
type CloudAuthService struct {
	cloudAPIURL string
	httpClient  *http.Client
}

// NewCloudAuthService creates a new cloud auth service
func NewCloudAuthService(cloudAPIURL string) *CloudAuthService {
	return &CloudAuthService{
		cloudAPIURL: strings.TrimRight(cloudAPIURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CloudValidateResponse represents the cloud auth validate response
type CloudValidateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Valid                 bool   `json:"valid"`
		UserID                string `json:"user_id"`
		Email                 string `json:"email"`
		FirstName             string `json:"first_name"`
		LastName              string `json:"last_name"`
		IsActive              bool   `json:"is_active"`
		SubscriptionStatus    string `json:"subscription_status"`
		SubscriptionExpiresAt string `json:"subscription_expires_at"`
		User                  struct {
			ID                    string `json:"id"`
			Email                 string `json:"email"`
			FirstName             string `json:"first_name"`
			LastName              string `json:"last_name"`
			IsActive              bool   `json:"is_active"`
			SubscriptionStatus    string `json:"subscription_status"`
			SubscriptionExpiresAt string `json:"subscription_expires_at"`
		} `json:"user"`
	} `json:"data"`
}

func (r *CloudValidateResponse) cloudUserID() string {
	if id := strings.TrimSpace(r.Data.User.ID); id != "" {
		return id
	}
	return strings.TrimSpace(r.Data.UserID)
}

func (r *CloudValidateResponse) cloudSubscriptionStatus() string {
	if status := strings.TrimSpace(r.Data.SubscriptionStatus); status != "" {
		return status
	}
	return strings.TrimSpace(r.Data.User.SubscriptionStatus)
}

func (r *CloudValidateResponse) cloudSubscriptionExpiresAt() string {
	if expires := strings.TrimSpace(r.Data.SubscriptionExpiresAt); expires != "" {
		return expires
	}
	return strings.TrimSpace(r.Data.User.SubscriptionExpiresAt)
}

// CloudSessionRequest represents a cloud session creation request
type CloudSessionRequest struct {
	CloudToken string `json:"cloud_token" binding:"required"`
}

// CloudSessionResponse represents a cloud session creation response
type CloudSessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// ValidateCloudToken validates a cloud access token against the Render API
func (s *CloudAuthService) ValidateCloudToken(ctx context.Context, cloudToken string) (*CloudValidateResponse, error) {
	url := fmt.Sprintf("%s/auth/validate", s.cloudAPIURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader("{}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create validate request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cloudToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call cloud validate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloud validation failed with status %d", resp.StatusCode)
	}

	var result CloudValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode cloud validate response: %w", err)
	}

	if !result.Data.Valid {
		return nil, fmt.Errorf("cloud token is not valid")
	}

	return &result, nil
}

// CreateLocalSession creates a local JWT session from a validated cloud token.
// Subscription permission comes only from the cloud validate response; local
// SQLite columns are copied for display and must not block a cloud-approved account.
func (s *CloudAuthService) CreateLocalSession(ctx context.Context, jwtService *JWTService, db *sqlx.DB, cloudToken string) (*CloudSessionResponse, error) {
	validation, err := s.ValidateCloudToken(ctx, cloudToken)
	if err != nil {
		return nil, fmt.Errorf("cloud token validation failed: %w", err)
	}

	userID, err := uuid.Parse(validation.cloudUserID())
	if err != nil {
		return nil, fmt.Errorf("invalid user ID from cloud: %w", err)
	}

	type sessionUserRow struct {
		ID                    string         `db:"id"`
		Email                 string         `db:"email"`
		FirstName             string         `db:"first_name"`
		LastName              string         `db:"last_name"`
		Phone                 sql.NullString `db:"phone"`
		IsActive              bool           `db:"is_active"`
		SubscriptionStatus    sql.NullString `db:"subscription_status"`
		SubscriptionExpiresAt sql.NullString `db:"subscription_expires_at"`
		CreatedAt             string         `db:"created_at"`
		UpdatedAt             string         `db:"updated_at"`
	}

	var row sessionUserRow
	err = db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&row.ID,
		&row.Email,
		&row.FirstName,
		&row.LastName,
		&row.Phone,
		&row.IsActive,
		&row.SubscriptionStatus,
		&row.SubscriptionExpiresAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found in local database: %w", err)
	}

	user, err := userFromRow(userRow{
		ID:                    row.ID,
		Email:                 row.Email,
		PasswordHash:          "",
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Phone:                 row.Phone,
		IsActive:              row.IsActive,
		LastLoginAt:           sql.NullString{},
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		SubscriptionStatus:    row.SubscriptionStatus,
		SubscriptionExpiresAt: row.SubscriptionExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse user row: %w", err)
	}

	if status := validation.cloudSubscriptionStatus(); status != "" {
		user.SubscriptionStatus = status
	}
	if expiresRaw := validation.cloudSubscriptionExpiresAt(); expiresRaw != "" {
		if parsed, parseErr := parseSQLiteTimestamp(expiresRaw); parseErr == nil {
			user.SubscriptionExpiresAt = &parsed
		}
	}
	user.IsActive = true

	accessToken, err := jwtService.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	now := time.Now()
	if status := validation.cloudSubscriptionStatus(); status != "" {
		_, err = db.ExecContext(ctx, `
			UPDATE users
			SET last_login_at = $1,
			    updated_at = $2,
			    is_active = $3,
			    subscription_status = $4,
			    subscription_expires_at = $5
			WHERE id = $6
		`, now, now, true, status, user.SubscriptionExpiresAt, user.ID)
	} else {
		_, err = db.ExecContext(ctx, "UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3", now, now, user.ID)
	}
	if err != nil {
		log.Printf("failed to update last login: %v", err)
	}

	return &CloudSessionResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		User:         user,
	}, nil
}
