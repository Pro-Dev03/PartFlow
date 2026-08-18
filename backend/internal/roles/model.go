package roles

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role
type Role struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	Permissions    []string  `json:"permissions" db:"permissions"` // Array of permission names
	IsSystem       bool      `json:"is_system" db:"is_system"`     // System roles cannot be deleted
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// RoleRequest represents role creation/update request
type RoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Define standard roles
const (
	RoleOwner    = "owner"
	RoleManager  = "manager"
	RoleEmployee = "employee"
	RoleAccountant = "accountant"
)

// GetStandardRoles returns standard role definitions
func GetStandardRoles() []Role {
	return []Role{
		{
			Name:        RoleOwner,
			Description: "Full access to all features",
			Permissions: []string{
				// Owner has all permissions
				"products.read", "products.create", "products.update", "products.delete", "products.archive",
				"inventory.read", "inventory.adjust", "inventory.transfer", "inventory.inspect",
				"sales.read", "sales.create", "sales.cancel", "sales.refund",
				"customers.read", "customers.create", "customers.update", "customers.delete",
				"debts.read", "debts.manage",
				"suppliers.read", "suppliers.create", "suppliers.update", "suppliers.delete",
				"purchases.read", "purchases.create", "purchases.receive",
				"expenses.read", "expenses.create", "expenses.update", "expenses.delete",
				"returns.read", "returns.create", "returns.approve",
				"warranties.read", "warranties.claim",
				"reports.read", "reports.export",
				"users.read", "users.create", "users.update", "users.delete",
				"settings.manage",
				"audit.read",
			},
			IsSystem: true,
		},
		{
			Name:        RoleManager,
			Description: "Manage operations and reports",
			Permissions: []string{
				"products.read", "products.create", "products.update", "products.archive",
				"inventory.read", "inventory.adjust", "inventory.transfer", "inventory.inspect",
				"sales.read", "sales.create", "sales.cancel",
				"customers.read", "customers.create", "customers.update",
				"debts.read", "debts.manage",
				"suppliers.read", "suppliers.create", "suppliers.update",
				"purchases.read", "purchases.create", "purchases.receive",
				"expenses.read", "expenses.create", "expenses.update",
				"returns.read", "returns.create", "returns.approve",
				"warranties.read",
				"reports.read", "reports.export",
				"users.read",
			},
			IsSystem: true,
		},
		{
			Name:        RoleEmployee,
			Description: "Sales and basic inventory operations",
			Permissions: []string{
				"products.read",
				"inventory.read", "inventory.transfer",
				"sales.read", "sales.create",
				"customers.read", "customers.create",
				"debts.read",
				"suppliers.read",
				"returns.read",
				"warranties.read",
			},
			IsSystem: true,
		},
		{
			Name:        RoleAccountant,
			Description: "Financial operations and reports",
			Permissions: []string{
				"sales.read",
				"customers.read",
				"debts.read", "debts.manage",
				"suppliers.read",
				"purchases.read",
				"expenses.read", "expenses.create", "expenses.update",
				"returns.read",
				"reports.read", "reports.export",
			},
			IsSystem: true,
		},
	}
}
