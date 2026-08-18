package roles

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateRole(ctx context.Context, req *CreateRoleRequest, organizationID uuid.UUID) (*Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	ListRoles(ctx context.Context, organizationID uuid.UUID, page, pageSize int, search string) ([]Role, int, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req *UpdateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	InitializeStandardRoles(ctx context.Context, organizationID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateRole(ctx context.Context, req *CreateRoleRequest, organizationID uuid.UUID) (*Role, error) {
	if req.Name == "" {
		return nil, ErrRoleNameRequired
	}

	// Check if role with same name already exists
	existing, err := s.repo.GetByName(ctx, organizationID, req.Name)
	if err == nil && existing != nil {
		return nil, ErrRoleAlreadyExists
	}

	role := &Role{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    req.Permissions,
		IsSystem:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *service) GetRole(ctx context.Context, id uuid.UUID) (*Role, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListRoles(ctx context.Context, organizationID uuid.UUID, page, pageSize int, search string) ([]Role, int, error) {
	offset := (page - 1) * pageSize

	roles, err := s.repo.List(ctx, organizationID, pageSize, offset, search)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx, organizationID, search)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, req *UpdateRoleRequest) (*Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsSystem {
		return nil, ErrRoleSystemDelete
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Permissions != nil {
		role.Permissions = req.Permissions
	}
	role.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return ErrRoleSystemDelete
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) InitializeStandardRoles(ctx context.Context, organizationID uuid.UUID) error {
	standardRoles := GetStandardRoles()

	for _, stdRole := range standardRoles {
		// Check if role already exists
		existing, err := s.repo.GetByName(ctx, organizationID, stdRole.Name)
		if err == nil && existing != nil {
			// Role exists, skip
			continue
		}

		// Create role
		role := &Role{
			ID:             uuid.New(),
			OrganizationID: organizationID,
			Name:           stdRole.Name,
			Description:    stdRole.Description,
			Permissions:    stdRole.Permissions,
			IsSystem:       stdRole.IsSystem,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := s.repo.Create(ctx, role); err != nil {
			return err
		}
	}

	return nil
}