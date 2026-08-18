export type UserRole = 'owner' | 'manager' | 'employee' | 'accountant';

export type Permission =
  | 'dashboard.view'
  | 'inventory.view'
  | 'inventory.create'
  | 'inventory.edit'
  | 'inventory.delete'
  | 'sales.view'
  | 'sales.create'
  | 'sales.edit'
  | 'sales.delete'
  | 'customers.view'
  | 'customers.create'
  | 'customers.edit'
  | 'customers.delete'
  | 'debts.view'
  | 'debts.record_payment'
  | 'suppliers.view'
  | 'suppliers.create'
  | 'suppliers.edit'
  | 'suppliers.delete'
  | 'purchases.view'
  | 'purchases.create'
  | 'purchases.edit'
  | 'purchases.delete'
  | 'expenses.view'
  | 'expenses.create'
  | 'expenses.edit'
  | 'expenses.delete'
  | 'expenses.approve'
  | 'returns.view'
  | 'returns.create'
  | 'returns.edit'
  | 'returns.approve'
  | 'warranties.view'
  | 'warranties.create'
  | 'warranties.edit'
  | 'reports.view'
  | 'reports.generate'
  | 'settings.view'
  | 'settings.edit'
  | 'users.view'
  | 'users.create'
  | 'users.edit'
  | 'users.delete';

const ROLE_PERMISSIONS: Record<UserRole, Permission[]> = {
  owner: [
    // Full access to everything
    'dashboard.view',
    'inventory.view', 'inventory.create', 'inventory.edit', 'inventory.delete',
    'sales.view', 'sales.create', 'sales.edit', 'sales.delete',
    'customers.view', 'customers.create', 'customers.edit', 'customers.delete',
    'debts.view', 'debts.record_payment',
    'suppliers.view', 'suppliers.create', 'suppliers.edit', 'suppliers.delete',
    'purchases.view', 'purchases.create', 'purchases.edit', 'purchases.delete',
    'expenses.view', 'expenses.create', 'expenses.edit', 'expenses.delete', 'expenses.approve',
    'returns.view', 'returns.create', 'returns.edit', 'returns.approve',
    'warranties.view', 'warranties.create', 'warranties.edit',
    'reports.view', 'reports.generate',
    'settings.view', 'settings.edit',
    'users.view', 'users.create', 'users.edit', 'users.delete',
  ],
  manager: [
    // Most permissions except user management
    'dashboard.view',
    'inventory.view', 'inventory.create', 'inventory.edit',
    'sales.view', 'sales.create', 'sales.edit',
    'customers.view', 'customers.create', 'customers.edit',
    'debts.view', 'debts.record_payment',
    'suppliers.view', 'suppliers.create', 'suppliers.edit',
    'purchases.view', 'purchases.create', 'purchases.edit',
    'expenses.view', 'expenses.create', 'expenses.edit', 'expenses.approve',
    'returns.view', 'returns.create', 'returns.edit', 'returns.approve',
    'warranties.view', 'warranties.create', 'warranties.edit',
    'reports.view', 'reports.generate',
    'settings.view', 'settings.edit',
  ],
  employee: [
    // Limited permissions for daily operations
    'dashboard.view',
    'inventory.view',
    'sales.view', 'sales.create',
    'customers.view', 'customers.create',
    'debts.view',
    'suppliers.view',
    'purchases.view',
    'expenses.view',
    'returns.view',
    'warranties.view',
    'reports.view',
  ],
  accountant: [
    // Financial permissions
    'dashboard.view',
    'inventory.view',
    'sales.view',
    'customers.view',
    'debts.view', 'debts.record_payment',
    'suppliers.view',
    'purchases.view',
    'expenses.view', 'expenses.create', 'expenses.edit', 'expenses.approve',
    'returns.view',
    'warranties.view',
    'reports.view', 'reports.generate',
    'settings.view',
  ],
};

export function hasPermission(userRole: UserRole, permission: Permission): boolean {
  const permissions = ROLE_PERMISSIONS[userRole] || [];
  return permissions.includes(permission);
}

export function hasAnyPermission(userRole: UserRole, permissions: Permission[]): boolean {
  return permissions.some(permission => hasPermission(userRole, permission));
}

export function hasAllPermissions(userRole: UserRole, permissions: Permission[]): boolean {
  return permissions.every(permission => hasPermission(userRole, permission));
}

export function getUserPermissions(userRole: UserRole): Permission[] {
  return ROLE_PERMISSIONS[userRole] || [];
}