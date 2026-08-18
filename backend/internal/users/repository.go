package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for users
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new users repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, phone, avatar_url, role_id, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.OrganizationID, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone, user.AvatarURL,
		user.RoleID, user.IsActive, user.IsVerified, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*User, error) {
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, avatar_url, 
		       role_id, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	
	if organizationID != uuid.Nil {
		query += " AND organization_id = $2"
		var user User
		err := r.db.GetContext(ctx, &user, query, id, organizationID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrUserNotFound
			}
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		return &user, nil
	}
	
	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, avatar_url, 
		       role_id, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// List retrieves users with pagination and filters
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, page, perPage int, search string, isActive *bool, roleID *uuid.UUID) ([]User, int, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, avatar_url, 
		       role_id, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users WHERE organization_id = $1
	`
	args := []interface{}{organizationID}
	argCount := 1

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argCount, argCount, argCount)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argCount += 2
	}

	if isActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *isActive)
	}

	if roleID != nil {
		argCount++
		query += fmt.Sprintf(" AND role_id = $%d", argCount)
		args = append(args, *roleID)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users WHERE organization_id = $1"
	countArgs := []interface{}{organizationID}
	countArgCount := 1

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", countArgCount, countArgCount, countArgCount)
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
		countArgCount += 2
	}

	if isActive != nil {
		countArgCount++
		countQuery += fmt.Sprintf(" AND is_active = $%d", countArgCount)
		countArgs = append(countArgs, *isActive)
	}

	if roleID != nil {
		countArgCount++
		countQuery += fmt.Sprintf(" AND role_id = $%d", countArgCount)
		countArgs = append(countArgs, *roleID)
	}

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	var users []User
	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// Update updates a user
func (r *Repository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, phone = $6, avatar_url = $7, 
		    role_id = $8, is_active = $9, updated_at = $10
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Phone, user.AvatarURL, user.RoleID, user.IsActive, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *Repository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt interface{}) error {
	query := `UPDATE users SET last_login_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, lastLoginAt)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Delete deletes a user
func (r *Repository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	if organizationID != uuid.Nil {
		query += " AND organization_id = $2"
		result, err := r.db.ExecContext(ctx, query, id, organizationID)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrUserNotFound
		}
		return nil
	}
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}