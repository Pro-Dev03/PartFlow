import { useMemo } from 'react'

export type UserRole = 'owner' | 'manager' | 'employee' | 'accountant'

export interface Permission {
  resource: string
  action: 'create' | 'read' | 'update' | 'delete'
}

const rolePermissions: Record<UserRole, Permission[]> = {
  owner: [
    // Full access to everything
    { resource: 'products', action: 'create' },
    { resource: 'products', action: 'read' },
    { resource: 'products', action: 'update' },
    { resource: 'products', action: 'delete' },
    { resource: 'customers', action: 'create' },
    { resource: 'customers', action: 'read' },
    { resource: 'customers', action: 'update' },
    { resource: 'customers', action: 'delete' },
    { resource: 'sales', action: 'create' },
    { resource: 'sales', action: 'read' },
    { resource: 'sales', action: 'update' },
    { resource: 'sales', action: 'delete' },
    { resource: 'debts', action: 'create' },
    { resource: 'debts', action: 'read' },
    { resource: 'debts', action: 'update' },
    { resource: 'debts', action: 'delete' },
    { resource: 'inventory', action: 'create' },
    { resource: 'inventory', action: 'read' },
    { resource: 'inventory', action: 'update' },
    { resource: 'inventory', action: 'delete' },
    { resource: 'purchases', action: 'create' },
    { resource: 'purchases', action: 'read' },
    { resource: 'purchases', action: 'update' },
    { resource: 'purchases', action: 'delete' },
    { resource: 'suppliers', action: 'create' },
    { resource: 'suppliers', action: 'read' },
    { resource: 'suppliers', action: 'update' },
    { resource: 'suppliers', action: 'delete' },
    { resource: 'expenses', action: 'create' },
    { resource: 'expenses', action: 'read' },
    { resource: 'expenses', action: 'update' },
    { resource: 'expenses', action: 'delete' },
    { resource: 'reports', action: 'read' },
    { resource: 'reports', action: 'export' },
    { resource: 'settings', action: 'read' },
    { resource: 'settings', action: 'update' },
    { resource: 'users', action: 'create' },
    { resource: 'users', action: 'read' },
    { resource: 'users', action: 'update' },
    { resource: 'users', action: 'delete' },
    { resource: 'audit', action: 'read' },
  ],
  manager: [
    // Inventory and sales management
    { resource: 'products', action: 'create' },
    { resource: 'products', action: 'read' },
    { resource: 'products', action: 'update' },
    { resource: 'products', action: 'delete' },
    { resource: 'customers', action: 'create' },
    { resource: 'customers', action: 'read' },
    { resource: 'customers', action: 'update' },
    { resource: 'sales', action: 'create' },
    { resource: 'sales', action: 'read' },
    { resource: 'sales', action: 'update' },
    { resource: 'debts', action: 'read' },
    { resource: 'debts', action: 'update' },
    { resource: 'inventory', action: 'create' },
    { resource: 'inventory', action: 'read' },
    { resource: 'inventory', action: 'update' },
    { resource: 'purchases', action: 'create' },
    { resource: 'purchases', action: 'read' },
    { resource: 'purchases', action: 'update' },
    { resource: 'suppliers', action: 'create' },
    { resource: 'suppliers', action: 'read' },
    { resource: 'suppliers', action: 'update' },
    { resource: 'expenses', action: 'create' },
    { resource: 'expenses', action: 'read' },
    { resource: 'reports', action: 'read' },
    { resource: 'reports', action: 'export' },
    { resource: 'settings', action: 'read' },
    { resource: 'users', action: 'read' },
  ],
  employee: [
    // Sales and basic operations
    { resource: 'products', action: 'read' },
    { resource: 'customers', action: 'create' },
    { resource: 'customers', action: 'read' },
    { resource: 'sales', action: 'create' },
    { resource: 'sales', action: 'read' },
    { resource: 'inventory', action: 'read' },
    { resource: 'inventory', action: 'update' },
  ],
  accountant: [
    // Financial reports and expenses
    { resource: 'sales', action: 'read' },
    { resource: 'debts', action: 'read' },
    { resource: 'debts', action: 'update' },
    { resource: 'purchases', action: 'read' },
    { resource: 'expenses', action: 'create' },
    { resource: 'expenses', action: 'read' },
    { resource: 'expenses', action: 'update' },
    { resource: 'reports', action: 'read' },
    { resource: 'reports', action: 'export' },
    { resource: 'suppliers', action: 'read' },
  ],
}

export function usePermissions(role: UserRole) {
  const permissions = useMemo(() => rolePermissions[role] || [], [role])

  const hasPermission = (resource: string, action: string): boolean => {
    return permissions.some(
      (p) => p.resource === resource && p.action === action
    )
  }

  const hasAnyPermission = (resource: string, actions: string[]): boolean => {
    return actions.some((action) => hasPermission(resource, action))
  }

  const hasAllPermissions = (resource: string, actions: string[]): boolean => {
    return actions.every((action) => hasPermission(resource, action))
  }

  const canCreate = (resource: string): boolean => hasPermission(resource, 'create')
  const canRead = (resource: string): boolean => hasPermission(resource, 'read')
  const canUpdate = (resource: string): boolean => hasPermission(resource, 'update')
  const canDelete = (resource: string): boolean => hasPermission(resource, 'delete')

  return {
    permissions,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    canCreate,
    canRead,
    canUpdate,
    canDelete,
  }
}
