package roles

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, organizationID uuid.UUID, name string) (*Role, error)
	List(ctx context.Context, organizationID uuid.UUID, limit, offset int, search string) ([]Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, organizationID uuid.UUID, search string) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, role *Role) error {
	query := `
		INSERT INTO roles (id, organization_id, name, description, permissions, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, organization_id, name, description, permissions, is_system, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		role.ID,
		role.OrganizationID,
		role.Name,
		role.Description,
		role.Permissions,
		role.IsSystem,
		role.CreatedAt,
		role.UpdatedAt,
	).Scan(
		&role.ID,
		&role.OrganizationID,
		&role.Name,
		&role.Description,
		&role.Permissions,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	query := `
		SELECT id, organization_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	var role Role
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (r *repository) GetByName(ctx context.Context, organizationID uuid.UUID, name string) (*Role, error) {
	query := `
		SELECT id, organization_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE organization_id = $1 AND name = $2
	`

	var role Role
	err := r.db.GetContext(ctx, &role, query, organizationID, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return &role, nil
}

func (r *repository) List(ctx context.Context, organizationID uuid.UUID, limit, offset int, search string) ([]Role, error) {
	query := `
		SELECT id, organization_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE organization_id = $1
	`

	args := []interface{}{organizationID}

	if search != "" {
		query += ` AND (name ILIKE $2 OR description ILIKE $2)`
		args = append(args, "%"+search+"%")
		query += fmt.Sprintf(" OFFSET $%d LIMIT $%d", len(args)+1, len(args)+2)
		args = append(args, offset, limit)
	} else {
		query += fmt.Sprintf(" OFFSET $%d LIMIT $%d", len(args)+1, len(args)+2)
		args = append(args, offset, limit)
	}

	var roles []Role
	err := r.db.SelectContext(ctx, &roles, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

func (r *repository) Update(ctx context.Context, role *Role) error {
	query := `
		UPDATE roles
		SET name = $2, description = $3, permissions = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, organization_id, name, description, permissions, is_system, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		role.ID,
		role.Name,
		role.Description,
		role.Permissions,
		role.UpdatedAt,
	).Scan(
		&role.ID,
		&role.OrganizationID,
		&role.Name,
		&role.Description,
		&role.Permissions,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = false`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrRoleNotFound
	}

	return nil
}

func (r *repository) Count(ctx context.Context, organizationID uuid.UUID, search string) (int, error) {
	query := `SELECT COUNT(*) FROM roles WHERE organization_id = $1`
	args := []interface{}{organizationID}

	if search != "" {
		query += ` AND (name ILIKE $2 OR description ILIKE $2)`
		args = append(args, "%"+search+"%")
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return count, nil
}