package users

import (
	"context"
	"fmt"
	"strings"
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
func (s *Service) CreateUser(ctx context.Context, email, passwordHash, firstName, lastName string, phone, avatarURL *string, isActive bool, subscriptionDays int) (*User, error) {
	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserEmailExists
	}

	user := NewUser(email, passwordHash, firstName, lastName)
	if subscriptionDays > 0 {
		expiresAt := time.Now().UTC().AddDate(0, 0, subscriptionDays)
		user.SubscriptionExpiresAt = &expiresAt
	}
	user.Phone = phone
	user.AvatarURL = avatarURL
	user.IsActive = isActive

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUsers retrieves users with pagination and filters
func (s *Service) ListUsers(ctx context.Context, page, perPage int, search string, isActive *bool) ([]User, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, page, perPage, search, isActive)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, email, passwordHash, firstName, lastName string, phone, avatarURL *string, isActive bool) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
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
	user.IsActive = isActive
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UpdateSubscription updates a user's subscription status and expiry.
func (s *Service) UpdateSubscription(ctx context.Context, id uuid.UUID, status string, expiresAt *time.Time) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.SubscriptionStatus = status
	if expiresAt != nil {
		expiresAtUTC := expiresAt.UTC()
		expiresAt = &expiresAtUTC
	}
	user.SubscriptionExpiresAt = expiresAt
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	return user, nil
}

// RenewSubscription extends a user's subscription by the given number of days.
func (s *Service) RenewSubscription(ctx context.Context, id uuid.UUID, days int) (*User, error) {
	if days <= 0 {
		return nil, fmt.Errorf("days must be greater than zero")
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	normalizedStatus := strings.ToLower(strings.TrimSpace(user.SubscriptionStatus))
	if user.SubscriptionExpiresAt == nil || normalizedStatus == "expired" || normalizedStatus == "canceled" || normalizedStatus == "cancelled" || normalizedStatus == "deleted" {
		newExpiry := time.Now().UTC().AddDate(0, 0, days)
		user.SubscriptionExpiresAt = &newExpiry
		user.SubscriptionStatus = "active"
	} else {
		newExpiry := user.SubscriptionExpiresAt.AddDate(0, 0, days)
		user.SubscriptionExpiresAt = &newExpiry
		if normalizedStatus != "active" {
			user.SubscriptionStatus = "active"
		}
	}
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to renew subscription: %w", err)
	}
	return user, nil
}

// GetSubscriptionSummary returns basic counts for the admin dashboard.
func (s *Service) GetSubscriptionSummary(ctx context.Context) (map[string]int, error) {
	users, _, err := s.repo.List(ctx, 1, 500, "", nil)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"total": 0, "active": 0, "expired": 0, "canceled": 0}
	for _, user := range users {
		counts["total"]++
		normalizedStatus := strings.ToLower(strings.TrimSpace(user.SubscriptionStatus))
		switch normalizedStatus {
		case "active":
			counts["active"]++
		case "expired":
			counts["expired"]++
		case "canceled", "cancelled", "deleted":
			counts["canceled"]++
		}
	}
	return counts, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
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
