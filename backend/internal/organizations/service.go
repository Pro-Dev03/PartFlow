package organizations

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles business logic for organizations
type Service struct {
	repo *Repository
}

// NewService creates a new organizations service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new organization
func (s *Service) Create(ctx context.Context, org *Organization) error {
	// Check if slug already exists
	existing, err := s.repo.GetBySlug(ctx, org.Slug)
	if err == nil && existing != nil {
		return fmt.Errorf("organization with slug '%s' already exists", org.Slug)
	}

	return s.repo.Create(ctx, org)
}

// GetByID retrieves an organization by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySlug retrieves an organization by slug
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// Update updates an organization
func (s *Service) Update(ctx context.Context, org *Organization) error {
	// Check if organization exists
	_, err := s.repo.GetByID(ctx, org.ID)
	if err != nil {
		return err
	}

	// Check if new slug conflicts with existing organization
	if org.Slug != "" {
		existing, err := s.repo.GetBySlug(ctx, org.Slug)
		if err == nil && existing != nil && existing.ID != org.ID {
			return fmt.Errorf("organization with slug '%s' already exists", org.Slug)
		}
	}

	return s.repo.Update(ctx, org)
}

// Delete deletes an organization
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// List retrieves all organizations with pagination
func (s *Service) List(ctx context.Context, page, perPage int) ([]*Organization, int, error) {
	offset := (page - 1) * perPage
	orgs, err := s.repo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}
