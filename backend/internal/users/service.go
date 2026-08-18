package users

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service handles business logic for users
type Service struct {
	repo *Repository
}

// NewService creates a new users service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, organizationID uuid.UUID, email, passwordHash, firstName, lastName string, phone, avatarURL *string, roleID *uuid.UUID, isActive bool) (*User, error) {
	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserEmailExists
	}

	user := NewUser(organizationID, email, passwordHash, firstName, lastName)
	user.Phone = phone
	user.AvatarURL = avatarURL
	user.RoleID = roleID
	user.IsActive = isActive

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id, organizationID)
}

// ListUsers retrieves users with pagination and filters
func (s *Service) ListUsers(ctx context.Context, organizationID uuid.UUID, page, perPage int, search string, isActive *bool, roleID *uuid.UUID) ([]User, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, organizationID, page, perPage, search, isActive, roleID)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, email, passwordHash, firstName, lastName string, phone, avatarURL *string, roleID *uuid.UUID, isActive bool) (*User, error) {
	user, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if new email conflicts with existing user
	if email != "" && email != user.Email {
		existing, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existing != nil {
			return nil, ErrUserEmailExists
		}
	}

	// Update fields
	if email != "" {
		user.Email = email
	}
	if passwordHash != "" {
		user.PasswordHash = passwordHash
	}
	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}
	user.Phone = phone
	user.AvatarURL = avatarURL
	user.RoleID = roleID
	user.IsActive = isActive
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.Delete(ctx, id, organizationID)
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetByID(ctx, userID, uuid.Nil) // No organization check for password change
	if err != nil {
		return err
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrCurrentPasswordIncorrect
	}

	// Validate new password length
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Authenticate authenticates a user with email and password
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// UpdateLastLogin updates the last login timestamp
func (s *Service) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return s.repo.UpdateLastLogin(ctx, userID, now)
}

// AssignRole assigns a role to a user
func (s *Service) AssignRole(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, roleID uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, userID, organizationID)
	if err != nil {
		return err
	}

	user.RoleID = &roleID
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user
func (s *Service) RemoveRole(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, roleID uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, userID, organizationID)
	if err != nil {
		return err
	}

	if user.RoleID == nil || *user.RoleID != roleID {
		return fmt.Errorf("user does not have this role")
	}

	user.RoleID = nil
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// GetUserRoles gets the roles assigned to a user
func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) ([]interface{}, error) {
	user, err := s.repo.GetByID(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}

	if user.RoleID == nil {
		return []interface{}{}, nil
	}

	// Return role information
	return []interface{}{
		map[string]interface{}{
			"role_id": user.RoleID,
		},
	}, nil
}