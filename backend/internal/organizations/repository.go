package organizations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for organizations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new organizations repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new organization
func (r *Repository) Create(ctx context.Context, org *Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, email, phone, address, city, country, logo_url, settings, subscription_plan, subscription_status, trial_ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		org.ID, org.Name, org.Slug, org.Email, org.Phone, org.Address,
		org.City, org.Country, org.LogoURL, org.Settings,
		org.SubscriptionPlan, org.SubscriptionStatus, org.TrialEndsAt,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

// GetByID retrieves an organization by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	query := `
		SELECT id, name, slug, email, phone, address, city, country, logo_url, settings, 
		       subscription_plan, subscription_status, trial_ends_at, created_at, updated_at
		FROM organizations WHERE id = $1
	`

	var org Organization
	err := r.db.GetContext(ctx, &org, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// GetBySlug retrieves an organization by slug
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	query := `
		SELECT id, name, slug, email, phone, address, city, country, logo_url, settings, 
		       subscription_plan, subscription_status, trial_ends_at, created_at, updated_at
		FROM organizations WHERE slug = $1
	`

	var org Organization
	err := r.db.GetContext(ctx, &org, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// Update updates an organization
func (r *Repository) Update(ctx context.Context, org *Organization) error {
	query := `
		UPDATE organizations 
		SET name = $2, slug = $3, email = $4, phone = $5, address = $6, city = $7, country = $8,
		    logo_url = $9, settings = $10, subscription_plan = $11, subscription_status = $12, trial_ends_at = $13
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		org.ID, org.Name, org.Slug, org.Email, org.Phone, org.Address,
		org.City, org.Country, org.LogoURL, org.Settings,
		org.SubscriptionPlan, org.SubscriptionStatus, org.TrialEndsAt,
	).Scan(&org.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// Delete deletes an organization
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// List retrieves all organizations with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Organization, error) {
	query := `
		SELECT id, name, slug, email, phone, address, city, country, logo_url, settings, 
		       subscription_plan, subscription_status, trial_ends_at, created_at, updated_at
		FROM organizations ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	var orgs []*Organization
	err := r.db.SelectContext(ctx, &orgs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, nil
}

// Count returns the total number of organizations
func (r *Repository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM organizations`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}
	return count, nil
}
