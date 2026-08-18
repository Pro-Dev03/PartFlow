package roles

import "errors"

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleNameRequired   = errors.New("role name is required")
	ErrRoleSystemDelete   = errors.New("system roles cannot be deleted")
	ErrRoleAlreadyExists = errors.New("role with this name already exists")
	ErrInvalidPermission  = errors.New("invalid permission")
)