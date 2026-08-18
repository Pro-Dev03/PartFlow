package permissions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// HasPermission checks if a user has a specific permission
func (s *Service) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM role_permissions rp
			JOIN roles r ON rp.role_id = r.id
			JOIN users u ON u.role_id = r.id
			WHERE u.id = $1 AND rp.permission_name = $2
		)
	`

	var hasPermission bool
	err := s.db.GetContext(ctx, &hasPermission, query, userID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *Service) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT rp.permission_name
		FROM role_permissions rp
		JOIN roles r ON rp.role_id = r.id
		JOIN users u ON u.role_id = r.id
		WHERE u.id = $1
		ORDER BY rp.permission_name
	`

	var permissions []string
	err := s.db.SelectContext(ctx, &permissions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role
func (s *Service) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT permission_name
		FROM role_permissions
		WHERE role_id = $1
		ORDER BY permission_name
	`

	var permissions []string
	err := s.db.SelectContext(ctx, &permissions, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *Service) AssignPermissionToRole(ctx context.Context, roleID uuid.UUID, permission string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_name, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (role_id, permission_name) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query, roleID, permission)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *Service) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission string) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_name = $2
	`

	_, err := s.db.ExecContext(ctx, query, roleID, permission)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
}

// InitializeStandardPermissions creates standard permissions in the database
func (s *Service) InitializeStandardPermissions(ctx context.Context) error {
	standardPermissions := GetStandardPermissions()

	for _, perm := range standardPermissions {
		query := `
			INSERT INTO permissions (id, name, description, resource, action, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (name) DO UPDATE SET
				description = EXCLUDED.description,
				resource = EXCLUDED.resource,
				action = EXCLUDED.action
		`

		_, err := s.db.ExecContext(ctx, query,
			uuid.New(), perm.Name, perm.Description, perm.Resource, perm.Action)
		if err != nil {
			return fmt.Errorf("failed to initialize permission %s: %w", perm.Name, err)
		}
	}

	return nil
}
