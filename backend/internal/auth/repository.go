package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles auth data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, phone, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.OrganizationID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.RoleID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, role_id, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, role_id, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// UpdateLastLogin updates the last login timestamp
func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $1, updated_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, userID)
	return err
}

// CreateRefreshToken creates a new refresh token
func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

// GetRefreshToken retrieves a refresh token by token string
func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`
	var refreshToken RefreshToken
	err := r.db.GetContext(ctx, &refreshToken, query, token)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidToken
	}
	return &refreshToken, err
}

// DeleteRefreshToken deletes a refresh token
func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// DeleteUserRefreshTokens deletes all refresh tokens for a user
func (r *Repository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// UpdatePassword updates user password
func (r *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, passwordHash, now, userID)
	return err
}

// GetRole retrieves a role by ID
func (r *Repository) GetRole(ctx context.Context, roleID uuid.UUID) (*Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = $1
	`
	var role Role
	err := r.db.GetContext(ctx, &role, query, roleID)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &role, err
}

// GetUserPermissions retrieves all permissions for a user
func (r *Repository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1
	`
	var permissions []Permission
	err := r.db.SelectContext(ctx, &permissions, query, userID)
	return permissions, err
}
