package roles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStandardRoles(t *testing.T) {
	t.Run("returns correct number of standard roles", func(t *testing.T) {
		roles := GetStandardRoles()

		assert.Equal(t, 4, len(roles))
	})

	t.Run("owner has all permissions", func(t *testing.T) {
		roles := GetStandardRoles()
		ownerRole := roles[0]

		assert.Equal(t, RoleOwner, ownerRole.Name)
		assert.True(t, ownerRole.IsSystem)
		assert.Greater(t, len(ownerRole.Permissions), 10)
		assert.Contains(t, ownerRole.Permissions, "products.read")
		assert.Contains(t, ownerRole.Permissions, "sales.create")
		assert.Contains(t, ownerRole.Permissions, "users.create")
		assert.Contains(t, ownerRole.Permissions, "users.delete")
	})

	t.Run("manager has operational permissions", func(t *testing.T) {
		roles := GetStandardRoles()
		managerRole := roles[1]

		assert.Equal(t, RoleManager, managerRole.Name)
		assert.True(t, managerRole.IsSystem)
		assert.Contains(t, managerRole.Permissions, "products.create")
		assert.Contains(t, managerRole.Permissions, "sales.create")
		assert.NotContains(t, managerRole.Permissions, "users.manage")
	})

	t.Run("employee has limited permissions", func(t *testing.T) {
		roles := GetStandardRoles()
		employeeRole := roles[2]

		assert.Equal(t, RoleEmployee, employeeRole.Name)
		assert.True(t, employeeRole.IsSystem)
		assert.LessOrEqual(t, len(employeeRole.Permissions), 12)
		assert.Contains(t, employeeRole.Permissions, "products.read")
		assert.Contains(t, employeeRole.Permissions, "sales.create")
		assert.NotContains(t, employeeRole.Permissions, "products.delete")
	})

	t.Run("accountant has financial permissions", func(t *testing.T) {
		roles := GetStandardRoles()
		accountantRole := roles[3]

		assert.Equal(t, RoleAccountant, accountantRole.Name)
		assert.True(t, accountantRole.IsSystem)
		assert.Contains(t, accountantRole.Permissions, "debts.manage")
		assert.Contains(t, accountantRole.Permissions, "expenses.create")
		assert.Contains(t, accountantRole.Permissions, "reports.read")
		assert.NotContains(t, accountantRole.Permissions, "sales.create")
	})
}

func TestRoleConstants(t *testing.T) {
	t.Run("role constants are defined", func(t *testing.T) {
		assert.Equal(t, "owner", RoleOwner)
		assert.Equal(t, "manager", RoleManager)
		assert.Equal(t, "employee", RoleEmployee)
		assert.Equal(t, "accountant", RoleAccountant)
	})
}